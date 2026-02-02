// ============================================
// BUDGETBAKERS AUTOMATION - PLAYWRIGHT VERSION
// ============================================
//
// SETUP:
// 1. Install Node.js from https://nodejs.org/
// 2. Run: npm install
// 3. Run: npm run setup  (installs Firefox browser)
// 4. Update APPS_SCRIPT_URL below if needed
// 5. Run: npm start
//
// FIRST RUN:
// - Browser will open, you'll need to login manually
// - After login, the script will save your session
// - Future runs will use saved session (no login needed)
//
// ============================================

const { chromium } = require('playwright');
const fs = require('fs');
const SHARED = require('./config');

// ============================================
// CONFIGURATION
// ============================================

// Parse command line arguments
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
Usage: node budgetbakers-transaction.js [options]

Options:
  --skip <n>     Number of records to skip (default: 0)
  --limit <n>    Number of records to fetch, 0 = all (default: 0)
  --mode <mode>  'auto' = auto submit, 'assist' = manual review (default: auto)
  --help, -h     Show this help

Examples:
  node budgetbakers-transaction.js --skip 100 --limit 10
  make transaction-auto SKIP=100 LIMIT=10
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

const stripSuffix = (value) => {
  if (!value) return '';
  return value.replace(/::\d+$/, '').trim();
};

// Fetch transactions from Google Sheets
async function fetchTransactions() {
  const params = new URLSearchParams();
  if (CONFIG.SKIP > 0) params.append('skip', CONFIG.SKIP);
  if (CONFIG.LIMIT > 0) params.append('limit', CONFIG.LIMIT);

  const url = `${CONFIG.APPS_SCRIPT_URL}?${params.toString()}`;
  console.log(`Fetching transactions (skip=${CONFIG.SKIP}, limit=${CONFIG.LIMIT})...`);

  const response = await fetch(url);
  const result = await response.json();

  const transactions = result.data || result;
  console.log(`Fetched ${transactions.length} transactions (total: ${result.meta?.total_data || transactions.length})`);

  return transactions;
}

// Update status in Google Sheet after successful insert
async function updateStatus(no, status = 'INSERTED') {
  try {
    const response = await fetch(CONFIG.APPS_SCRIPT_URL, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ no, status }),
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

// ============================================
// MAIN AUTOMATION
// ============================================

// Helper to wait for user input
async function waitForEnter(prompt = 'Press ENTER to continue...') {
  console.log(prompt);
  await new Promise(resolve => {
    process.stdin.resume();
    process.stdin.once('data', () => {
      process.stdin.pause();
      resolve();
    });
  });
}

async function runAutomation() {
  console.log('='.repeat(50));
  console.log('BUDGETBAKERS PLAYWRIGHT AUTOMATION');
  console.log(`Mode: ${CONFIG.MODE.toUpperCase()}`);
  console.log(`Skip: ${CONFIG.SKIP}, Limit: ${CONFIG.LIMIT}`);
  console.log('='.repeat(50));

  // Fetch transactions first
  let transactions;
  try {
    transactions = await fetchTransactions();
  } catch (error) {
    console.error('Failed to fetch transactions:', error.message);
    return;
  }

  if (!transactions || transactions.length === 0) {
    console.log('No transactions to process');
    return;
  }

  console.log(`Processing ${transactions.length} transactions...\n`);

  // Launch browser (Firefox works better with Google OAuth)
  console.log('Launching browser...');
  const browser = await chromium.launch({
    headless: CONFIG.HEADLESS,
    slowMo: CONFIG.SLOW_MO,
  });
  console.log('Browser launched.');

  // Create context with saved session if available
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

  // Set long timeouts to allow manual login
  page.setDefaultTimeout(300000); // 5 minutes
  page.setDefaultNavigationTimeout(300000); // 5 minutes

  try {
    // Navigate to BudgetBakers
    console.log('Opening BudgetBakers...');
    await page.goto('https://web.budgetbakers.com/dashboard', {
      waitUntil: 'domcontentloaded', // Don't wait for network idle
      timeout: 120000,
    });

    // Wait a bit for redirects
    await delay(3000);

    // Check if we need to login
    const currentUrl = page.url();
    if (!currentUrl.includes('/dashboard') || currentUrl.includes('login') || currentUrl.includes('auth')) {
      console.log('\n' + '='.repeat(60));
      console.log('>>> MANUAL LOGIN REQUIRED <<<');
      console.log('='.repeat(60));
      console.log('');
      console.log('Option 1: Login with Google (if not blocked)');
      console.log('');
      console.log('Option 2: Login with Email');
      console.log('  - Enter your email in the Playwright browser');
      console.log('  - Check your email for the login link');
      console.log('  - IMPORTANT: Copy the link from email');
      console.log('  - Paste it into the PLAYWRIGHT browser address bar');
      console.log('  - (Do NOT click the link - it opens wrong browser)');
      console.log('');
      console.log('After you see the dashboard, press ENTER here to continue');
      console.log('='.repeat(60) + '\n');

      // Wait for user to press Enter (no timeout)
      await new Promise(resolve => {
        process.stdin.resume();
        process.stdin.once('data', () => {
          process.stdin.pause();
          resolve();
        });
      });

      // Wait for dashboard to be visible
      console.log('Checking dashboard...');
      await page.waitForSelector('button:has-text("Record")', { timeout: 180000 });

      // Save session for future use
      await context.storageState({ path: CONFIG.SESSION_PATH });
      console.log('Session saved for future runs.\n');
    }

    // Wait for dashboard to load
    await page.waitForSelector('button:has-text("Record")', { timeout: 180000 });
    console.log('Dashboard loaded.\n');

    // Process each transaction
    let successCount = 0;
    let failCount = 0;

    for (let i = 0; i < transactions.length; i++) {
      const tx = transactions[i];
      const account = stripSuffix(tx.account);
      const category = stripSuffix(tx.category);
      const amount = tx.amount;
      const note = tx.note || '';
      const type = (tx.type || 'expense').toLowerCase(); // 'expense', 'income', or 'transfer'

      console.log(`[${i + 1}/${transactions.length}] ${account} - Rp${amount} (${type})`);

      try {
        // Click Record button to open modal
        console.log('  Opening modal...');
        await page.click('button:has-text("Record")');
        await page.waitForSelector('[role="dialog"]', { timeout: 10000 });
        await delay(500);
        console.log('  Modal opened.');

        // Select transaction type (Expense/Income/Transfer)
        // Always click explicitly - don't assume default selection
        const radioGroup = page.locator('[role="radiogroup"]');
        if (type === 'income') {
          await radioGroup.getByText('Income', { exact: true }).click();
          await delay(CONFIG.DELAY_AFTER_FIELD);
          console.log(`  Type: Income`);
        } else if (type === 'transfer') {
          await radioGroup.getByText('Transfer', { exact: true }).click();
          await delay(CONFIG.DELAY_AFTER_FIELD);
          console.log(`  Type: Transfer`);
        } else {
          // Explicitly select Expense (don't assume it's selected)
          await radioGroup.getByText('Expense', { exact: true }).click();
          await delay(CONFIG.DELAY_AFTER_FIELD);
          console.log(`  Type: Expense`);
        }

        // Fill Amount
        const amountInput = page.getByRole('textbox', { name: 'Amount' });
        await amountInput.fill(String(Math.abs(parseFloat(amount))));
        await delay(CONFIG.DELAY_AFTER_FIELD);
        console.log(`  Amount: ${Math.abs(parseFloat(amount))}`);

        // Select Account (click to open dropdown, scroll to find option)
        const accountInput = page.getByRole('textbox', { name: 'Account' });
        await accountInput.click();
        await delay(500);

        // Wait for dropdown to appear
        await page.waitForSelector('[role="option"]', { timeout: 5000 }).catch(() => { });
        const accOptions = page.locator('[role="option"]');
        const accCount = await accOptions.count();
        let accBestMatch = null;
        let accBestIndex = -1;

        // Find exact match by scrolling through options
        for (let o = 0; o < accCount; o++) {
          const option = accOptions.nth(o);
          const optionText = await option.textContent({ timeout: 2000 }).catch(() => '');
          const trimmed = optionText ? optionText.trim() : '';

          if (trimmed === account) {
            accBestMatch = trimmed;
            accBestIndex = o;
            break;
          }

          if (trimmed.includes(account) && (!accBestMatch || trimmed.length < accBestMatch.length)) {
            accBestMatch = trimmed;
            accBestIndex = o;
          }
        }

        if (accBestIndex >= 0) {
          const selectedOption = accOptions.nth(accBestIndex);
          await selectedOption.scrollIntoViewIfNeeded();
          await delay(100);
          await selectedOption.click({ timeout: 5000 });
        } else if (accCount > 0) {
          await accOptions.first().click({ timeout: 5000 });
        } else {
          await page.keyboard.press('Escape');
        }
        await delay(CONFIG.DELAY_AFTER_FIELD);
        console.log(`  Account: ${account}`);

        // Select Category
        const categoryInput = page.getByRole('textbox', { name: 'Category' });
        await categoryInput.click();
        await delay(300);

        // Type to filter and find exact match
        await page.keyboard.type(category, { delay: 30 });
        await delay(500);

        // Wait for dropdown, then find exact match
        await page.waitForSelector('[role="option"]', { timeout: 5000 }).catch(() => { });
        const options = page.locator('[role="option"]');
        const count = await options.count();
        let bestMatch = null;
        let bestMatchIndex = -1;

        for (let o = 0; o < count; o++) {
          const optionText = await options.nth(o).textContent({ timeout: 2000 }).catch(() => '');
          const trimmed = optionText ? optionText.trim() : '';

          // Exact match - use immediately
          if (trimmed === category) {
            bestMatch = trimmed;
            bestMatchIndex = o;
            break;
          }

          // Track shortest containing match as fallback
          if (trimmed.includes(category) && (!bestMatch || trimmed.length < bestMatch.length)) {
            bestMatch = trimmed;
            bestMatchIndex = o;
          }
        }

        if (bestMatchIndex >= 0) {
          await options.nth(bestMatchIndex).scrollIntoViewIfNeeded();
          await delay(100);
          await options.nth(bestMatchIndex).click({ timeout: 5000 });
          if (bestMatch === category) {
            console.log(`  Category: ${category}`);
          } else {
            console.log(`  Category: ${category} -> ${bestMatch}`);
          }
        } else if (count > 0) {
          await options.first().click({ timeout: 5000 });
          console.log(`  Category: ${category} (not found, used first)`);
        } else {
          await page.keyboard.press('Escape');
          console.log(`  Category: ${category} (dropdown empty)`);
        }
        await delay(CONFIG.DELAY_AFTER_FIELD);

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

          // Wait for dropdown, then find exact match
          await page.waitForSelector('[role="option"]', { timeout: 5000 }).catch(() => { });
          const labelOptions = page.locator('[role="option"]');
          const labelCount = await labelOptions.count();
          let labelFound = false;

          for (let o = 0; o < labelCount; o++) {
            const optionText = await labelOptions.nth(o).textContent({ timeout: 2000 }).catch(() => '');
            const trimmed = optionText ? optionText.trim() : '';

            if (trimmed === 'generated') {
              await labelOptions.nth(o).scrollIntoViewIfNeeded();
              await delay(100);
              await labelOptions.nth(o).click({ timeout: 5000 });
              labelFound = true;
              break;
            }
          }

          // Fallback: click first if no exact match
          if (!labelFound && labelCount > 0) {
            await labelOptions.first().click({ timeout: 5000 });
          }

          await delay(200);

          // Close dropdown
          await page.keyboard.press('Escape');
          await delay(CONFIG.DELAY_AFTER_FIELD);
          console.log(`  Label: generated`);
        } catch (e) {
          console.log(`  Label: Could not add`);
        }

        // Set Date (if provided)
        if (tx.date) {
          console.log(`  Setting date: ${tx.date}`);
          // Parse date (format: DD/MM/YYYY)
          const [day, month, year] = tx.date.split('/');
          const targetDay = parseInt(day);
          const targetMonth = parseInt(month);
          const targetYear = parseInt(year);

          // Month names for building aria-label
          const monthNames = ['January', 'February', 'March', 'April', 'May', 'June',
            'July', 'August', 'September', 'October', 'November', 'December'];

          // Click on the Date & Time button to open picker
          await page.locator('button').filter({ hasText: /\d{1,2}\/\d{1,2}\/\d{4}/ }).first().click();
          await delay(500);

          // Get current displayed month/year from the calendar header button
          const headerButton = page.locator('.mantine-DateTimePicker-calendarHeaderLevel, button').filter({ hasText: /[A-Za-z]+ \d{4}/ }).first();
          const headerText = await headerButton.textContent().catch(() => '');
          console.log(`  Calendar shows: ${headerText}`);

          // Parse current month from header (e.g., "November 2025")
          let currentMonth = new Date().getMonth() + 1;
          let currentYear = new Date().getFullYear();

          for (let i = 0; i < monthNames.length; i++) {
            if (headerText.includes(monthNames[i])) {
              currentMonth = i + 1;
              const yearMatch = headerText.match(/\d{4}/);
              if (yearMatch) currentYear = parseInt(yearMatch[0]);
              break;
            }
          }

          // Calculate months to navigate
          const monthDiff = (targetYear - currentYear) * 12 + (targetMonth - currentMonth);
          console.log(`  Need to navigate ${monthDiff} months`);

          // Navigate months using Mantine DateTimePicker control buttons
          // First button is previous, second is next
          const navButtons = page.locator('.mantine-DateTimePicker-calendarHeaderControl');

          if (monthDiff < 0) {
            const prevButton = navButtons.first();
            for (let m = 0; m < Math.abs(monthDiff); m++) {
              await prevButton.click();
              await delay(200);
              console.log(`  Navigated back ${m + 1} month(s)`);
            }
          } else if (monthDiff > 0) {
            const nextButton = navButtons.nth(1);
            for (let m = 0; m < monthDiff; m++) {
              await nextButton.click();
              await delay(200);
              console.log(`  Navigated forward ${m + 1} month(s)`);
            }
          }

          await delay(300);

          // Click the target day using aria-label (e.g., "1 September 2025")
          const targetMonthName = monthNames[targetMonth - 1];
          const dayLabel = `${targetDay} ${targetMonthName} ${targetYear}`;
          console.log(`  Clicking day with label: ${dayLabel}`);

          await page.getByLabel(dayLabel, { exact: true }).click();
          await delay(200);

          // Close the date picker by pressing Escape or clicking outside
          await page.keyboard.press('Escape');
          await delay(CONFIG.DELAY_AFTER_FIELD);
          console.log(`  Date set to: ${tx.date}`);
        }

        // ASSIST MODE: Wait for user to click submit in browser
        if (CONFIG.MODE === 'assist') {
          console.log('\n  >> ASSIST MODE: Review the form and click submit in browser');
          console.log('  >> Or press ENTER here after submitting manually');

          // Store current amount to detect form reset
          const currentAmount = await page.getByRole('textbox', { name: 'Amount' }).inputValue().catch(() => '');

          // Wait for either: user presses Enter OR form amount changes (indicating submit)
          const submitted = await Promise.race([
            // Wait for user to press Enter in terminal
            new Promise(resolve => {
              process.stdin.resume();
              process.stdin.once('data', () => {
                process.stdin.pause();
                resolve('enter');
              });
            }),
            // Or detect form reset (amount becomes empty or changes significantly)
            (async () => {
              while (true) {
                await delay(500);
                const newAmount = await page.getByRole('textbox', { name: 'Amount' }).inputValue().catch(() => currentAmount);
                // If amount is now empty or very different, form was likely reset
                if (newAmount === '' || (currentAmount !== '' && newAmount !== currentAmount && Math.abs(parseFloat(newAmount.replace(/[^0-9.-]/g, '') || 0)) < 10)) {
                  return 'form_reset';
                }
                // Check if modal closed
                const modalVisible = await page.locator('[role="dialog"]').isVisible().catch(() => false);
                if (!modalVisible) {
                  return 'modal_closed';
                }
              }
            })(),
          ]);

          console.log(`  ✓ SUBMITTED (detected: ${submitted})`);
          successCount++;

          // Update status in Google Sheet
          await updateStatus(tx.no, 'INSERTED');
          console.log('');

          // If there are more transactions, re-open modal if needed
          if (i < transactions.length - 1) {
            await delay(500);
            const modalVisible = await page.locator('[role="dialog"]').isVisible().catch(() => false);
            if (!modalVisible) {
              console.log('  Re-opening modal for next transaction...');
              await page.click('button:has-text("Record")');
              await page.waitForSelector('[role="dialog"]', { timeout: 10000 });
              await delay(500);
            }
          }

          continue; // Skip the auto-submit code below
        }

        // AUTO MODE: Click submit button automatically
        const submitButton = page.getByRole('button', { name: 'Add and create another' });
        if (await submitButton.isVisible()) {
          await submitButton.click();
        } else {
          await page.getByRole('button', { name: 'Add record' }).click();
        }

        await delay(CONFIG.DELAY_AFTER_SUBMIT);
        console.log(`  ✓ SUCCESS`);
        successCount++;

        // Update status in Google Sheet
        await updateStatus(tx.no, 'INSERTED');
        console.log('');

        // Wait before next transaction
        if (i < transactions.length - 1) {
          await delay(CONFIG.DELAY_BETWEEN_TRANSACTIONS);
        }

      } catch (error) {
        // Sound alert (beep) and visual warning
        process.stdout.write('\x07'); // Terminal beep
        console.error('\n' + '!'.repeat(50));
        console.error(`  ✗ FAILED: ${error.message}`);
        console.error('!'.repeat(50) + '\n');
        failCount++;

        // Try to close modal if open
        try {
          const closeButton = page.locator('[role="dialog"] button').first();
          if (await closeButton.isVisible()) {
            await closeButton.click();
            await delay(500);
          }
        } catch { }
      }
    }

    // Summary
    console.log('='.repeat(50));
    console.log('AUTOMATION COMPLETE');
    console.log(`Success: ${successCount}/${transactions.length}`);
    console.log(`Failed: ${failCount}/${transactions.length}`);
    console.log('='.repeat(50));

    // Save session at the end
    await context.storageState({ path: CONFIG.SESSION_PATH });

  } catch (error) {
    console.error('Automation error:', error.message);
  } finally {
    // Close browser
    await browser.close();
  }
}

// Run the automation
runAutomation().catch(console.error);
