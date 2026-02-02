// ============================================
// BUDGETBAKERS AUTOMATION - TRANSFER SCRIPT
// ============================================
//
// Handles transfer transactions from the Transfer sheet
// Uses same session as transaction script
//
// Usage:
//   node budgetbakers-transfer.js --skip 0 --limit 10 --mode auto
//   make transfer-auto SKIP=0 LIMIT=10
//
// ============================================

const { chromium } = require('playwright');
const fs = require('fs');
const SHARED = require('./config');

// ============================================
// CONFIGURATION
// ============================================

function parseArgs() {
  const args = process.argv.slice(2);
  const parsed = {
    skip: 0,
    limit: 0,
    mode: 'auto',
  };

  for (let i = 0; i < args.length; i++) {
    const arg = args[i].toLowerCase();
    if (arg === '--skip' && args[i + 1]) {
      parsed.skip = parseInt(args[++i]) || 0;
    } else if (arg === '--limit' && args[i + 1]) {
      parsed.limit = parseInt(args[++i]) || 0;
    } else if (arg === '--mode' && args[i + 1]) {
      parsed.mode = args[++i].toLowerCase();
    } else if (arg === '--help' || arg === '-h') {
      console.log(`
Usage: node budgetbakers-transfer.js [options]

Options:
  --skip <n>     Number of records to skip (default: 0)
  --limit <n>    Number of records to fetch, 0 = all (default: 0)
  --mode <mode>  'auto' = auto submit, 'assist' = manual review (default: auto)
  --help, -h     Show this help

Examples:
  node budgetbakers-transfer.js --skip 0 --limit 10
  make transfer-auto SKIP=0 LIMIT=10
`);
      process.exit(0);
    }
  }

  return parsed;
}

const ARGS = parseArgs();

const CONFIG = {
  ...SHARED,
  SKIP: ARGS.skip,
  LIMIT: ARGS.limit,
  MODE: ARGS.mode,
};

// ============================================
// HELPER FUNCTIONS
// ============================================

const delay = (ms) => new Promise(resolve => setTimeout(resolve, ms));

// Fetch transfers from Google Sheets
async function fetchTransfers() {
  const params = new URLSearchParams();
  params.append('type', 'transfer');
  if (CONFIG.SKIP > 0) params.append('skip', CONFIG.SKIP);
  if (CONFIG.LIMIT > 0) params.append('limit', CONFIG.LIMIT);

  const url = `${CONFIG.APPS_SCRIPT_URL}?${params.toString()}`;
  console.log(`Fetching transfers (skip=${CONFIG.SKIP}, limit=${CONFIG.LIMIT})...`);

  const response = await fetch(url);
  const result = await response.json();

  const transfers = result.data || result;
  console.log(`Fetched ${transfers.length} transfers (total: ${result.meta?.total_data || transfers.length})`);

  return transfers;
}

// Update status in Google Sheet
async function updateStatus(no, status = 'INSERTED') {
  try {
    const response = await fetch(CONFIG.APPS_SCRIPT_URL, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ no, status, type: 'transfer' }),
      redirect: 'follow',
    });
    const result = await response.json();
    if (result.success) {
      console.log(`  Sheet updated: ${status}`);
    } else {
      console.log(`  Sheet update failed: ${result.error}`);
    }
    return result.success;
  } catch (error) {
    console.log(`  Sheet update error: ${error.message}`);
    return false;
  }
}

// Select account from dropdown using getByRole (more reliable)
async function selectAccount(page, inputLocator, accountName, label) {
  await inputLocator.click();
  await delay(500);

  // Wait for listbox to appear
  await page.waitForSelector('[role="listbox"]', { timeout: 5000 }).catch(() => { });
  await delay(200);

  // Try to click option directly using getByRole (only matches visible elements)
  // Options have format: "Cash icon Cash Eko", "General icon DANA", etc.
  const option = page.getByRole('option').filter({ hasText: accountName });
  const count = await option.count();

  if (count > 0) {
    await option.first().click({ timeout: 5000 });
    console.log(`  ${label}: ${accountName}`);
  } else {
    // Fallback: click first visible option
    const firstOption = page.getByRole('option').first();
    if (await firstOption.isVisible().catch(() => false)) {
      await firstOption.click({ timeout: 5000 });
      console.log(`  ${label}: ${accountName} (not found, used first)`);
    } else {
      await page.keyboard.press('Escape');
      console.log(`  ${label}: ${accountName} (dropdown empty)`);
    }
  }
  await delay(CONFIG.DELAY_AFTER_FIELD);
}

// ============================================
// MAIN AUTOMATION
// ============================================

async function runAutomation() {
  console.log('='.repeat(50));
  console.log('BUDGETBAKERS TRANSFER AUTOMATION');
  console.log(`Mode: ${CONFIG.MODE.toUpperCase()}`);
  console.log(`Skip: ${CONFIG.SKIP}, Limit: ${CONFIG.LIMIT}`);
  console.log('='.repeat(50));

  let transfers;
  try {
    transfers = await fetchTransfers();
  } catch (error) {
    console.error('Failed to fetch transfers:', error.message);
    return;
  }

  if (!transfers || transfers.length === 0) {
    console.log('No transfers to process');
    return;
  }

  console.log(`Processing ${transfers.length} transfers...\n`);

  console.log('Launching browser...');
  const browser = await chromium.launch({
    headless: CONFIG.HEADLESS,
    slowMo: CONFIG.SLOW_MO,
  });
  console.log('Browser launched.');

  let context;
  if (fs.existsSync(CONFIG.SESSION_PATH)) {
    console.log('Loading saved session...');
    try {
      context = await browser.newContext({
        storageState: CONFIG.SESSION_PATH,
      });
      console.log('Session loaded.');
    } catch (e) {
      console.log('Session invalid, starting fresh...');
      fs.unlinkSync(CONFIG.SESSION_PATH);
      context = await browser.newContext();
    }
  } else {
    console.log('No saved session found. You will need to login.');
    context = await browser.newContext();
  }

  const page = await context.newPage();
  page.setDefaultTimeout(300000);
  page.setDefaultNavigationTimeout(300000);

  try {
    console.log('Opening BudgetBakers...');
    await page.goto('https://web.budgetbakers.com/dashboard', {
      waitUntil: 'domcontentloaded',
      timeout: 120000,
    });

    await delay(3000);

    const currentUrl = page.url();
    if (!currentUrl.includes('/dashboard') || currentUrl.includes('login') || currentUrl.includes('auth')) {
      console.log('\n' + '='.repeat(60));
      console.log('>>> MANUAL LOGIN REQUIRED <<<');
      console.log('='.repeat(60));
      console.log('');
      console.log('Please login in the browser, then press ENTER here');
      console.log('='.repeat(60) + '\n');

      await new Promise(resolve => {
        process.stdin.resume();
        process.stdin.once('data', () => {
          process.stdin.pause();
          resolve();
        });
      });

      console.log('Checking dashboard...');
      await page.waitForSelector('button:has-text("Record")', { timeout: 180000 });
      await context.storageState({ path: CONFIG.SESSION_PATH });
      console.log('Session saved for future runs.\n');
    }

    await page.waitForSelector('button:has-text("Record")', { timeout: 180000 });
    console.log('Dashboard loaded.\n');

    let successCount = 0;
    let failCount = 0;

    for (let i = 0; i < transfers.length; i++) {
      const tx = transfers[i];
      const fromAccount = tx.from_account;
      const toAccount = tx.to_account;
      const amount = tx.amount;
      const note = tx.note || '';

      console.log(`[${i + 1}/${transfers.length}] ${fromAccount} -> ${toAccount} - Rp${amount}`);

      try {
        // Open modal
        console.log('  Opening modal...');
        await page.click('button:has-text("Record")');
        await page.waitForSelector('[role="dialog"]', { timeout: 10000 });
        await delay(500);
        console.log('  Modal opened.');

        // Select Transfer type
        const radioGroup = page.locator('[role="radiogroup"]');
        await radioGroup.getByText('Transfer', { exact: true }).click();
        await delay(CONFIG.DELAY_AFTER_FIELD);
        console.log('  Type: Transfer');

        // Fill Amount (use getByLabel to avoid matching disabled Amount field)
        const amountInput = page.getByLabel('Amount *');
        await amountInput.fill(String(Math.abs(parseFloat(amount))));
        await delay(CONFIG.DELAY_AFTER_FIELD);
        console.log(`  Amount: ${Math.abs(parseFloat(amount))}`);

        // Select From Account
        const fromAccountInput = page.getByRole('textbox', { name: 'From account' });
        await selectAccount(page, fromAccountInput, fromAccount, 'From');

        // Select To Account (dropdown closes automatically after selecting From)
        const toAccountInput = page.getByRole('textbox', { name: 'To account' });
        await selectAccount(page, toAccountInput, toAccount, 'To');

        // Fill Note
        if (note) {
          const noteInput = page.getByRole('textbox', { name: 'Note' });
          await noteInput.fill(note);
          await delay(CONFIG.DELAY_AFTER_FIELD);
          console.log(`  Note: ${note.substring(0, 30)}${note.length > 30 ? '...' : ''}`);
        }

        // Add "generated" label
        try {
          const labelsInput = page.getByRole('textbox', { name: 'Labels' });
          await labelsInput.click();
          await delay(300);

          // Type to filter the dropdown
          await page.keyboard.type('generated', { delay: 50 });
          await delay(500);

          // Click option inside VISIBLE listbox (avoids hidden options from previous dropdowns)
          await page.waitForSelector('[role="listbox"]:visible', { timeout: 5000 }).catch(() => { });
          await page.locator('[role="listbox"]:visible [role="option"]').first().click({ timeout: 5000 });

          await delay(200);

          // Close dropdown
          await page.keyboard.press('Escape');
          await delay(CONFIG.DELAY_AFTER_FIELD);
          console.log(`  Label: generated`);
        } catch (e) {
          console.log(`  Label: Could not add (${e.message})`);
        }

        // Set Date
        if (tx.date) {
          console.log(`  Setting date: ${tx.date}`);
          const [day, month, year] = tx.date.split('/');
          const targetDay = parseInt(day);
          const targetMonth = parseInt(month);
          const targetYear = parseInt(year);

          const monthNames = ['January', 'February', 'March', 'April', 'May', 'June',
            'July', 'August', 'September', 'October', 'November', 'December'];

          await page.locator('button').filter({ hasText: /\d{1,2}\/\d{1,2}\/\d{4}/ }).first().click();
          await delay(500);

          const headerButton = page.locator('.mantine-DateTimePicker-calendarHeaderLevel, button').filter({ hasText: /[A-Za-z]+ \d{4}/ }).first();
          const headerText = await headerButton.textContent().catch(() => '');
          console.log(`  Calendar shows: ${headerText}`);

          let currentMonth = new Date().getMonth() + 1;
          let currentYear = new Date().getFullYear();

          for (let m = 0; m < monthNames.length; m++) {
            if (headerText.includes(monthNames[m])) {
              currentMonth = m + 1;
              const yearMatch = headerText.match(/\d{4}/);
              if (yearMatch) currentYear = parseInt(yearMatch[0]);
              break;
            }
          }

          const monthDiff = (targetYear - currentYear) * 12 + (targetMonth - currentMonth);
          console.log(`  Need to navigate ${monthDiff} months`);

          const navButtons = page.locator('.mantine-DateTimePicker-calendarHeaderControl');

          if (monthDiff < 0) {
            const prevButton = navButtons.first();
            for (let m = 0; m < Math.abs(monthDiff); m++) {
              await prevButton.click();
              await delay(200);
            }
          } else if (monthDiff > 0) {
            const nextButton = navButtons.nth(1);
            for (let m = 0; m < monthDiff; m++) {
              await nextButton.click();
              await delay(200);
            }
          }

          await delay(300);

          const targetMonthName = monthNames[targetMonth - 1];
          const dayLabel = `${targetDay} ${targetMonthName} ${targetYear}`;
          console.log(`  Clicking day with label: ${dayLabel}`);

          await page.getByLabel(dayLabel, { exact: true }).click();
          await delay(200);

          await page.keyboard.press('Escape');
          await delay(CONFIG.DELAY_AFTER_FIELD);
          console.log(`  Date set to: ${tx.date}`);
        }

        // ASSIST MODE
        if (CONFIG.MODE === 'assist') {
          console.log('\n  >> ASSIST MODE: Review and click submit in browser');
          console.log('  >> Or press ENTER here after submitting manually');

          const currentAmount = await page.getByLabel('Amount *').inputValue().catch(() => '');

          const submitted = await Promise.race([
            new Promise(resolve => {
              process.stdin.resume();
              process.stdin.once('data', () => {
                process.stdin.pause();
                resolve('enter');
              });
            }),
            (async () => {
              while (true) {
                await delay(500);
                const newAmount = await page.getByLabel('Amount *').inputValue().catch(() => currentAmount);
                if (newAmount === '' || (currentAmount !== '' && newAmount !== currentAmount)) {
                  return 'form_reset';
                }
                const modalVisible = await page.locator('[role="dialog"]').isVisible().catch(() => false);
                if (!modalVisible) {
                  return 'modal_closed';
                }
              }
            })(),
          ]);

          console.log(`  ✓ SUBMITTED (detected: ${submitted})`);
          successCount++;
          await updateStatus(tx.no, 'INSERTED');
          console.log('');

          if (i < transfers.length - 1) {
            await delay(500);
            const modalVisible = await page.locator('[role="dialog"]').isVisible().catch(() => false);
            if (!modalVisible) {
              console.log('  Re-opening modal for next transfer...');
              await page.click('button:has-text("Record")');
              await page.waitForSelector('[role="dialog"]', { timeout: 10000 });
              await delay(500);
            }
          }

          continue;
        }

        // AUTO MODE
        const submitButton = page.getByRole('button', { name: 'Add and create another' });
        if (await submitButton.isVisible()) {
          await submitButton.click();
        } else {
          await page.getByRole('button', { name: 'Add record' }).click();
        }

        await delay(CONFIG.DELAY_AFTER_SUBMIT);
        console.log(`  ✓ SUCCESS`);
        successCount++;

        await updateStatus(tx.no, 'INSERTED');
        console.log('');

        if (i < transfers.length - 1) {
          await delay(CONFIG.DELAY_BETWEEN_TRANSACTIONS);
        }

      } catch (error) {
        process.stdout.write('\x07');
        console.error('\n' + '!'.repeat(50));
        console.error(`  ✗ FAILED: ${error.message}`);
        console.error('!'.repeat(50) + '\n');
        failCount++;

        try {
          const closeButton = page.locator('[role="dialog"] button').first();
          if (await closeButton.isVisible()) {
            await closeButton.click();
            await delay(500);
          }
        } catch { }
      }
    }

    console.log('='.repeat(50));
    console.log('TRANSFER AUTOMATION COMPLETE');
    console.log(`Success: ${successCount}/${transfers.length}`);
    console.log(`Failed: ${failCount}/${transfers.length}`);
    console.log('='.repeat(50));

    await context.storageState({ path: CONFIG.SESSION_PATH });

  } catch (error) {
    console.error('Automation error:', error.message);
  } finally {
    await browser.close();
  }
}

runAutomation().catch(console.error);
