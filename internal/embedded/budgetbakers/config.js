// ============================================
// SHARED CONFIGURATION
// ============================================
// Environment variables are passed from Go binary or .env file
// ============================================

const path = require('path');

// Try to load dotenv if available (for standalone usage)
try {
  const dotenv = require('dotenv');
  // Try multiple .env locations
  const envPaths = [
    path.join(__dirname, '.env'),                    // Same directory (when extracted by Go)
    path.join(__dirname, '..', '..', '.env'),        // Project root (when running from scripts/budgetbakers)
    path.join(process.cwd(), '.env'),                // Current working directory
  ];
  for (const envPath of envPaths) {
    const result = dotenv.config({ path: envPath });
    if (!result.error) break;
  }
} catch (e) {
  // dotenv not available, rely on environment variables passed by Go
}

// Validate required env vars
if (!process.env.APPS_SCRIPT_URL) {
  console.error('ERROR: APPS_SCRIPT_URL environment variable is required');
  console.error('Set it via .env file or pass it as environment variable');
  process.exit(1);
}

// Session path - use env var if provided (Go will set this), otherwise use __dirname
const sessionPath = process.env.BB_SESSION_PATH || path.join(__dirname, '.budgetbakers-session');

module.exports = {
  // Google Apps Script URL (from env)
  APPS_SCRIPT_URL: process.env.APPS_SCRIPT_URL,

  // Browser settings (can be overridden via env)
  HEADLESS: process.env.BB_HEADLESS === 'true' || false,
  SLOW_MO: parseInt(process.env.BB_SLOW_MO) || 100,

  // Session storage path
  SESSION_PATH: sessionPath,

  // Delays (in milliseconds)
  DELAY_BETWEEN_TRANSACTIONS: parseInt(process.env.BB_DELAY_BETWEEN) || 2000,
  DELAY_AFTER_SUBMIT: parseInt(process.env.BB_DELAY_SUBMIT) || 1500,
  DELAY_AFTER_FIELD: parseInt(process.env.BB_DELAY_FIELD) || 300,
};
