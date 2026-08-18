# Label Handling Plan

Status: **menunggu approval** — belum ada kode yang ditulis.
Revisi 1: BudgetMate tetap hard fail (behaviour existing), nama field `labelIds`
sudah dikonfirmasi benar.

Menambahkan dukungan label (kolom I di sheet `Transaction`) agar ikut terkirim ke
BudgetMate dan BudgetBakers.

## Keputusan yang sudah disepakati

| Topik | Keputusan |
|---|---|
| Delimiter multi-label | Koma — `"BUSINESS, ADJUSTMENT"` |
| Field BudgetBakers | `labelIds` (array of id) — sudah dikonfirmasi benar |
| Field BudgetMate | `Labels` (array of id) |
| Label tidak ditemukan | Ikuti pola existing masing-masing app: BudgetBakers skip, BudgetMate hard fail |
| Scope | Hanya sheet `Transaction`. `Transfer` tidak disentuh |

## Temuan penting saat telaah kode

**1. Ada satu perubahan wajib di JS yang belum disebut: `google-apps-script.js:84`.**

```js
const data = sheet.getRange(5, 2, lastRow - 4, 7).getValues(); // B5:H
```

Angka `7` = jumlah kolom yang dibaca (B..H). Kolom I tidak ikut terbaca, jadi
`row[7]` akan selalu `undefined` dan label akan **selalu kosong tanpa error apa
pun**. Harus jadi `8` (B..I). Ini bagian yang paling mudah terlewat karena tidak
memunculkan pesan gagal.

**2. Pola error handling kedua app berbeda, dan itu dipertahankan.**

- `budgetbakers_app/sync_transaction.go` → `continue` (skip) saat account/category tidak ketemu.
- `budgetmate_app/sync_transaction.go` → `return fmt.Errorf(...)` (hard fail).

Label mengikuti pola masing-masing app, jadi tidak ada perubahan alur error di
BudgetMate. Efek sampingnya justru menguntungkan: karena BudgetMate berhenti di
transaksi pertama yang gagal, tidak ada risiko `createdNos` terisi transaksi yang
tidak pernah terkirim, sehingga status di Google Sheet tetap konsisten.

**3. Perlu normalisasi case.** Sheet memakai UPPERCASE (`BUSINESS`), API
mengembalikan Title Case (`Business`). Matching harus case-insensitive, mengikuti
pola `GetCategoryIDWithMapping` yang sudah dipakai di repo.

**4. BudgetBakers punya flag `archived`.** Label yang sudah diarsipkan harus
di-skip saat membangun lookup map, agar label lama tidak salah termatch.

## Rencana perubahan

### A. `scripts/budgetbakers/google-apps-script.js`

1. **Baris 84** — baca 8 kolom, bukan 7:
   ```js
   const data = sheet.getRange(5, 2, lastRow - 4, 8).getValues(); // B5:I
   ```

2. **Baris ~96-103** — tambah `labels` pada objek transaction:
   ```js
   // Columns: No(0) Date(1) Account(2) Category(3) Amount(4) Note(5) Type(6) Label(7)
   labels: parseLabels(row[7])
   ```

3. **Helper baru** (sebelah `stripSuffix`):
   ```js
   // "BUSINESS, ADJUSTMENT" -> ["BUSINESS","ADJUSTMENT"]; cell kosong -> []
   function parseLabels(value) {
     if (!value) return [];
     return String(value).split(',')
       .map(function (s) { return s.trim(); })
       .filter(function (s) { return s.length > 0; });
   }
   ```
   Dikirim sebagai **array**, bukan string, supaya Go tidak perlu mengulang
   parsing delimiter dan sumber kebenaran delimiter cuma di satu tempat.

4. **Baris 232** — geser kolom status transaction (label kini menempati I):
   ```js
   'transaction': { sheetName: 'Transaction', statusColumns: { 'budgetbakers': 10, 'budgetmate': 11 }, ... },
   ```
   Baris 233 (`transfer`) **tidak diubah**.

5. **Header dokumentasi (baris ~17-34)** — perbarui peta kolom agar tidak
   menyesatkan pembaca berikutnya: `I: Label`, status pindah ke `J`/`K`.

Peta kolom sesudah perubahan:

| Kolom | Isi |
|---|---|
| B–H | No, Date, Account, Category, Amount, Note, Type |
| I | **Label** (multi, dipisah koma) |
| J | Status BudgetBakers (col 10) |
| K | Status BudgetMate (col 11) |

### B. `internal/adapters/secondary/appsscript/types.go`

Tambah satu field pada `SheetTransaction`:
```go
Labels []string `json:"labels"`
```
`SheetTransfer` tidak diubah. Field lain tetap `string` mengikuti gaya existing;
`Labels` memakai slice karena memang multi-value.

### C. BudgetMate (`internal/adapters/secondary/budgetmateapi/`)

- **`types.go`** — tambah:
  ```go
  type Label struct {
      ID    string `json:"id"`
      Name  string `json:"name"`
      Color string `json:"color,omitempty"`
  }
  type LabelsResponse struct {
      Success bool    `json:"success"`
      Data    []Label `json:"data"`
  }
  ```
  `CreateTransactionRequest.Labels []string` **sudah ada** — tidak perlu diubah,
  hanya perlu diisi.

- **`label.go` (baru)** — `FetchLabels(ctx)` GET `/api/v1/labels`, mengisi
  `labelMap[lower(name)] = id`. Mengikuti pola `category.go` termasuk fallback
  ke array langsung bila response tidak ter-wrap. Plus `GetLabelIDWithMapping`
  dan `GetLabelNames` (untuk pesan error yang informatif).

- **`client.go`** — tambah `labelMap` dan `labelMapping` pada struct `Client`
  dan inisialisasinya di `NewClient`.

- **`mapping.go`** — `SetLabelMapping` mengikuti pola yang sudah ada.

### D. BudgetBakers (`internal/adapters/secondary/budgetbakers/`)

- **`models.go`** — tambah struct `Label` (`id`, `name`, `color`, `archived`)
  dan satu field pada `Record`:
  ```go
  LabelIDs []string `json:"labelIds,omitempty"`
  ```
  `omitempty` penting supaya payload transaksi tanpa label tidak berubah bentuk
  dari perilaku sekarang.

- **`get_labels.go` (baru)** — `GetLabels()` GET `/v1/api/labels?limit=200&offset=N`,
  pagination via `nextOffset` mengikuti `get_categories.go`, dan **melewati
  label `archived`**.

### E. Alur sync

- **`budgetbakers_app/adapter.go`** — tambah cache `labels []budgetbakers.Label`.
- **`budgetbakers_app/sync_transaction.go`** — muat labels bersama accounts &
  categories; resolve `tx.Labels` → `LabelIDs`; `continue` + warning bila ada
  label yang tidak dikenal (sejalan dengan pola existing di file ini).
- **`budgetmate_app/sync_transaction.go`** — `FetchLabels` saat inisialisasi;
  resolve `tx.Labels` → `createReq.Labels`; bila ada label yang tidak dikenal →
  `return fmt.Errorf(...)` menyebut nomor transaksi dan daftar label yang
  tersedia, **mengikuti persis pola `account not found` / `category not found`
  yang sudah ada**. Tidak ada perubahan alur error di file ini.

Resolusi label dibuat *all-or-nothing per transaksi*: satu label tidak dikenal →
seluruh transaksi tidak dikirim (di-skip di BudgetBakers, hard fail di
BudgetMate). Lebih aman daripada mengirim sebagian label, karena data yang masuk
separuh benar lebih sulit dideteksi daripada data yang jelas belum masuk.

## Rencana verifikasi

1. **Unit test** untuk resolusi label: case-insensitive match, label `archived`
   tidak terpakai, cell kosong → tanpa label, label tidak dikenal → skip
   (BudgetBakers) / error (BudgetMate).
2. **Test regresi**: transaksi tanpa label harus menghasilkan payload yang sama
   seperti sekarang (memastikan `omitempty` bekerja).
3. `go build ./...`, `go vet ./...`, `go test ./...`.
4. **Uji manual dry-run** dengan `LIMIT=1` sebelum menjalankan batch penuh,
   karena data terkirim ke layanan pihak ketiga dan tidak mudah dibatalkan.

## Risiko

| Risiko | Mitigasi |
|---|---|
| Kolom status lama di I belum dimigrasi ke J/K | Konfirmasi sheet sudah dirapikan sebelum deploy Apps Script |
| Apps Script dan binary Go tidak deploy bersamaan | Field `labels` bersifat aditif: Go lama mengabaikannya, Go baru melihat `[]` — jadi tidak ada crash di urutan deploy mana pun |
| Nama label di sheet berbeda ejaan dengan yang ada di app tujuan | BudgetBakers memberi warning lalu lanjut; BudgetMate berhenti dengan pesan berisi daftar label yang tersedia sehingga selisihnya langsung terlihat |

Meski nama field `labelIds` sudah dikonfirmasi, uji `LIMIT=1` tetap dianjurkan
sekali sebelum batch penuh — hanya untuk memastikan label benar-benar tampil di
UI, karena data terkirim ke layanan pihak ketiga dan tidak mudah dibatalkan.

## Di luar scope

- Sheet `Transfer` (belum punya kolom label).
- Membuat label baru otomatis bila belum ada di app tujuan.
- Auto-mapping nama label yang berbeda antar app (`SetLabelMapping` disiapkan
  tapi belum diisi).
