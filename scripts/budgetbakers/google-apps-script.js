// ============================================
// GOOGLE APPS SCRIPT FOR BUDGETBAKERS AUTOMATION
// ============================================
// 
// SETUP INSTRUCTIONS:
// 1. Open your Google Sheet with transaction data
// 2. Go to Extensions > Apps Script
// 3. Delete any existing code and paste this entire script
// 4. Click "Deploy" > "New deployment"
// 5. Select type: "Web app"
// 6. Set "Execute as": "Me"
// 7. Set "Who has access": "Anyone" (or "Anyone with Google account")
// 8. Click "Deploy" and authorize when prompted
// 9. Copy the Web app URL and paste it into budgetbakers-automation.js
//
// EXPECTED SHEET FORMAT (starting from B5):
// Column B: No (ignored)
// Column C: Date (date in DD/MM/YYYY format)
// Column D: Account (text, e.g., "OVO Dewi", "Cash Eko")
// Column E: Category (text, must match BudgetBakers categories)
// Column F: Amount (number, positive value)
// Column G: Note (text, description of transaction)
// Column H: Type (text, "income" or "expense")
//
// Data starts from row 5 (no header row)
// Sheet name must be "transaction"
//
// API USAGE:
// GET {url}                    - Returns all transactions
// GET {url}?skip=0&limit=50    - Returns paginated response (skip = offset)
//
// POST {url}                   - Update status after insert
//   Body: { "no": "123", "status": "INSERTED" }
//   Updates column I for the row with matching No
//
// Response format:
//   {
//     "data": [...],
//     "meta": {
//       "total_data": 150,
//       "total_filtered": 50,
//       "skip": 0,
//       "limit": 50
//     }
//   }
//
// ============================================

function doGet(e) {
  try {
    // Parse query parameters
    const params = e ? e.parameter : {};
    const skip = parseInt(params.skip) || 0;
    const limit = parseInt(params.limit) || 0; // 0 = no limit (return all)
    const type = params.type || 'transaction'; // 'transaction' or 'transfer'

    if (type === 'transfer') {
      return getTransfers(skip, limit);
    } else {
      return getTransactions(skip, limit);
    }

  } catch (error) {
    return ContentService
      .createTextOutput(JSON.stringify({ error: error.message }))
      .setMimeType(ContentService.MimeType.JSON);
  }
}

// Get transactions from Transaction sheet
function getTransactions(skip, limit) {
  const sheet = SpreadsheetApp.getActiveSpreadsheet().getSheetByName('Transaction');
  if (!sheet) {
    throw new Error('Sheet "Transaction" not found');
  }
  const lastRow = sheet.getLastRow();

  if (lastRow < 5) {
    return ContentService
      .createTextOutput(JSON.stringify({ data: [], meta: { total_data: 0 } }))
      .setMimeType(ContentService.MimeType.JSON);
  }

  const data = sheet.getRange(5, 2, lastRow - 4, 7).getValues(); // B5:H

  // Build transactions array
  // Columns: No(0), Date(1), Account(2), Category(3), Amount(4), Note(5), Type(6)
  const allTransactions = [];
  for (let i = 0; i < data.length; i++) {
    const row = data[i];

    // Skip empty rows (check Amount and Account)
    if (!row[4] && !row[2]) continue;

    const transaction = {
      no: String(row[0] || '0'),
      amount: String(row[4] || '0'),
      account: stripSuffix(String(row[2] || '')),
      category: stripSuffix(String(row[3] || '')),
      note: String(row[5] || ''),
      date: formatDate(row[1]),
      type: String(row[6] || 'expense').toLowerCase()
    };

    // Only add if has required fields
    if (transaction.amount && transaction.account) {
      allTransactions.push(transaction);
    }
  }

  const total = allTransactions.length;
  const startIndex = skip;
  const endIndex = limit > 0 ? startIndex + limit : total;
  const paginatedData = allTransactions.slice(startIndex, endIndex);

  return ContentService
    .createTextOutput(JSON.stringify({
      data: paginatedData,
      meta: { total_data: total, total_filtered: paginatedData.length, skip: skip, limit: limit }
    }))
    .setMimeType(ContentService.MimeType.JSON);
}

// Get transfers from Transfer sheet
function getTransfers(skip, limit) {
  const sheet = SpreadsheetApp.getActiveSpreadsheet().getSheetByName('Transfer');
  if (!sheet) {
    throw new Error('Sheet "Transfer" not found');
  }
  const lastRow = sheet.getLastRow();

  if (lastRow < 5) {
    return ContentService
      .createTextOutput(JSON.stringify({ data: [], meta: { total_data: 0 } }))
      .setMimeType(ContentService.MimeType.JSON);
  }

  const data = sheet.getRange(5, 2, lastRow - 4, 6).getValues(); // B5:G

  // Build transfers array
  // Columns: No(0), Date(1), From Account(2), To Account(3), Amount(4), Note(5)
  const allTransfers = [];
  for (let i = 0; i < data.length; i++) {
    const row = data[i];

    // Skip empty rows (check Amount and From Account)
    if (!row[4] && !row[2]) continue;

    const transfer = {
      no: String(row[0] || '0'),
      date: formatDate(row[1]),
      from_account: stripSuffix(String(row[2] || '')),
      to_account: stripSuffix(String(row[3] || '')),
      amount: parseAmount(row[4]),
      note: String(row[5] || '')
    };

    // Only add if has required fields
    if (transfer.amount && transfer.from_account && transfer.to_account) {
      allTransfers.push(transfer);
    }
  }

  const total = allTransfers.length;
  const startIndex = skip;
  const endIndex = limit > 0 ? startIndex + limit : total;
  const paginatedData = allTransfers.slice(startIndex, endIndex);

  return ContentService
    .createTextOutput(JSON.stringify({
      data: paginatedData,
      meta: { total_data: total, total_filtered: paginatedData.length, skip: skip, limit: limit }
    }))
    .setMimeType(ContentService.MimeType.JSON);
}

// Parse amount from Indonesian format (Rp1.234.567,89) to number string
function parseAmount(value) {
  if (!value) return '0';
  if (typeof value === 'number') return String(value);
  
  // Remove "Rp" prefix and thousands separators, convert decimal comma to dot
  let str = String(value)
    .replace(/Rp\s*/gi, '')
    .replace(/\./g, '')      // Remove thousands separator (.)
    .replace(/,/g, '.');     // Convert decimal separator (, to .)
  
  const num = parseFloat(str) || 0;
  return String(num);
}

// Remove suffix like ::1, ::2 from account/category names
function stripSuffix(value) {
  if (!value) return '';
  return value.replace(/::\d+$/, '').trim();
}

function formatDate(dateValue) {
  if (!dateValue) return '';

  // If already a string in DD/MM/YYYY format
  if (typeof dateValue === 'string') {
    return dateValue;
  }

  // If it's a Date object
  if (dateValue instanceof Date) {
    const day = String(dateValue.getDate()).padStart(2, '0');
    const month = String(dateValue.getMonth() + 1).padStart(2, '0');
    const year = dateValue.getFullYear();
    return `${day}/${month}/${year}`;
  }

  return String(dateValue);
}

// Handle POST requests to update status
function doPost(e) {
  try {
    const data = JSON.parse(e.postData.contents);
    const isBatch = Array.isArray(data);
    const updates = isBatch ? data : [data];
    
    if (updates.length === 0) {
      throw new Error('No data provided');
    }

    const results = [];
    
    // Group sheets logic to avoid redundant SpreadsheetApp calls
    const sheetsData = {
      'transaction': { sheetName: 'Transaction', statusColumn: 9, noColumn: null, lastRow: null, sheet: null },
      'transfer': { sheetName: 'Transfer', statusColumn: 8, noColumn: null, lastRow: null, sheet: null }
    };
    
    const getSheetData = (type) => {
      const typeKey = type || 'transaction';
      const sData = sheetsData[typeKey];
      
      if (!sData) {
         throw new Error(`Unsupported type: ${typeKey}`);
      }
      
      if (!sData.sheet) {
         sData.sheet = SpreadsheetApp.getActiveSpreadsheet().getSheetByName(sData.sheetName);
         if (!sData.sheet) {
           throw new Error(`Sheet "${sData.sheetName}" not found`);
         }
         sData.lastRow = sData.sheet.getLastRow();
         if (sData.lastRow >= 5) {
           sData.noColumn = sData.sheet.getRange(5, 2, sData.lastRow - 4, 1).getValues();
         } else {
           sData.noColumn = [];
         }
      }
      return sData;
    };
    
    for (const update of updates) {
      const no = String(update.no);
      const status = update.status || 'INSERTED';
      const type = update.type || 'transaction';
      
      if (!no) {
        results.push({ success: false, error: 'Missing "no" field' });
        continue;
      }
      
      try {
        const sData = getSheetData(type);
        
        let targetRow = -1;
        for (let i = 0; i < sData.noColumn.length; i++) {
          if (String(sData.noColumn[i][0]) === no) {
            targetRow = i + 5;
            break;
          }
        }
        
        if (targetRow === -1) {
          throw new Error(`Row with No "${no}" not found`);
        }
        
        sData.sheet.getRange(targetRow, sData.statusColumn).setValue(status);
        results.push({ success: true, no: no, targetRow: targetRow, type: type });
      } catch (err) {
        results.push({ success: false, no: no, error: err.message, type: type });
      }
    }
    
    return ContentService
      .createTextOutput(JSON.stringify({ 
        success: true, 
        message: isBatch ? `Processed ${updates.length} updates` : `Updated row ${results[0]?.targetRow} with status "${updates[0]?.status}"`,
        results: results
      }))
      .setMimeType(ContentService.MimeType.JSON);
      
  } catch (error) {
    return ContentService
      .createTextOutput(JSON.stringify({ 
        success: false, 
        error: error.message 
      }))
      .setMimeType(ContentService.MimeType.JSON);
  }
}

// Optional: Test function to verify data
function testGetTransactions() {
  const result = doGet();
  Logger.log(result.getContent());
}
