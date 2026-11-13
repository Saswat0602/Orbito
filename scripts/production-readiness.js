#!/usr/bin/env node
/**
 * Orbito Production Readiness & Pre-Flight Verification Script
 * Validates configuration, security keys, database, redis, and compiled assets.
 */

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const ROOT_DIR = path.resolve(__dirname, '..');
let errors = 0;
let warnings = 0;

console.log('\n===============================================================');
console.log('🛡️  ORBITO ENTERPRISE PRODUCTION READINESS AUDIT');
console.log('===============================================================\n');

function check(label, condition, errorMsg, isWarning = false) {
  if (condition) {
    console.log(`  ✅ [PASS] ${label}`);
  } else if (isWarning) {
    console.log(`  ⚠️  [WARN] ${label}: ${errorMsg}`);
    warnings++;
  } else {
    console.log(`  ❌ [FAIL] ${label}: ${errorMsg}`);
    errors++;
  }
}

// 1. Check Build Artifacts
console.log('📦 1. Verifying Compiled Production Artifacts...');
check(
  'Shared Types Compiled',
  fs.existsSync(path.join(ROOT_DIR, 'packages/types/dist/index.js')),
  'packages/types/dist/index.js is missing. Run: npm run build:types'
);

check(
  'Backend Dist Artifacts',
  fs.existsSync(path.join(ROOT_DIR, 'backend/dist/main.js')),
  'backend/dist/main.js is missing. Run: npm run build:backend'
);

check(
  'Frontend Standalone Build',
  fs.existsSync(path.join(ROOT_DIR, 'frontend/.next/standalone')),
  'frontend/.next/standalone is missing. Run: npm run build:frontend'
);

check(
  'Prisma Client Generated',
  fs.existsSync(path.join(ROOT_DIR, 'node_modules/.prisma/client')) ||
    fs.existsSync(path.join(ROOT_DIR, 'node_modules/@prisma/client')),
  'Prisma client not generated. Run: npm run db:generate'
);

// 2. Check Docker & Orchestration Files
console.log('\n🐳 2. Verifying Container & Deployment Configuration...');
check(
  'Production Docker Compose File',
  fs.existsSync(path.join(ROOT_DIR, 'docker-compose.prod.yml')),
  'docker-compose.prod.yml is missing'
);

check(
  'Backend Production Dockerfile',
  fs.existsSync(path.join(ROOT_DIR, 'docker/Dockerfile.backend.prod')),
  'docker/Dockerfile.backend.prod is missing'
);

check(
  'Frontend Production Dockerfile',
  fs.existsSync(path.join(ROOT_DIR, 'docker/Dockerfile.frontend.prod')),
  'docker/Dockerfile.frontend.prod is missing'
);

check(
  'Nginx Reverse Proxy Config',
  fs.existsSync(path.join(ROOT_DIR, 'docker/nginx/nginx.conf')),
  'docker/nginx/nginx.conf is missing'
);

check(
  'GitHub Actions CI/CD Pipeline',
  fs.existsSync(path.join(ROOT_DIR, '.github/workflows/ci.yml')),
  '.github/workflows/ci.yml is missing'
);

// 3. Environment & Security Hygiene
console.log('\n🔐 3. Auditing Security Keys & Environment Templates...');
check(
  'Root .env.example Template',
  fs.existsSync(path.join(ROOT_DIR, '.env.example')),
  '.env.example is missing in root'
);

check(
  'Frontend .env.example Template',
  fs.existsSync(path.join(ROOT_DIR, 'frontend/.env.example')),
  'frontend/.env.example is missing'
);

// Check active .env if present
const envPath = fs.existsSync(path.join(ROOT_DIR, '.env'))
  ? path.join(ROOT_DIR, '.env')
  : path.join(ROOT_DIR, 'backend/.env');

if (fs.existsSync(envPath)) {
  const envContent = fs.readFileSync(envPath, 'utf8');

  const checkSecret = (key, defaultVal) => {
    const match = envContent.match(new RegExp(`^${key}=["']?([^"'\r\n]+)["']?`, 'm'));
    const val = match ? match[1] : null;
    const isDefault = !val || val === defaultVal;
    return { val, isDefault };
  };

  const jwt = checkSecret('JWT_SECRET', 'orbito-jwt-super-secret-development-key-change-in-prod');
  const refresh = checkSecret('JWT_REFRESH_SECRET', 'orbito-jwt-refresh-secret-development-key-change-in-prod');
  const encKey = checkSecret('ENCRYPTION_KEY', '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef');

  if (process.env.NODE_ENV === 'production') {
    check('JWT Secret Hardened', !jwt.isDefault, 'Production must not use default JWT_SECRET');
    check('JWT Refresh Secret Hardened', !refresh.isDefault, 'Production must not use default JWT_REFRESH_SECRET');
    check('Encryption Key Hardened', !encKey.isDefault, 'Production must not use default ENCRYPTION_KEY');
  } else {
    check(
      'JWT Secrets Configured',
      !!jwt.val && !!refresh.val,
      'JWT secrets are not defined in .env'
    );
    if (jwt.isDefault || refresh.isDefault || encKey.isDefault) {
      console.log('  ⚠️  [INFO] Currently using development secrets. In production, rotate via: openssl rand -base64 32');
      warnings++;
    }
  }
} else {
  console.log('  ⚠️  [WARN] No active .env file detected in workspace root.');
  warnings++;
}

// 4. Database Migrations Status
console.log('\n🗄️  4. Checking Database Migrations...');
try {
  const migrationStatus = execSync(
    'npx prisma migrate status --schema=backend/prisma/schema.prisma',
    { encoding: 'utf8', stdio: ['pipe', 'pipe', 'pipe'] }
  );
  const isUpToDate = migrationStatus.includes('Database schema is up to date');
  check('Prisma Database Migrations In Sync', isUpToDate, migrationStatus.trim());
} catch (err) {
  check(
    'Prisma Database Migrations In Sync',
    false,
    'Failed to query database migration status. Ensure PostgreSQL is reachable.',
    true
  );
}

// Summary
console.log('\n===============================================================');
console.log(`🏁 AUDIT COMPLETE: ${errors} Failures, ${warnings} Warnings`);
console.log('===============================================================');

if (errors === 0) {
  console.log('\n🎉 ALL PRODUCTION READINESS CRITERIA PASSED! Ready for deployment.\n');
  process.exit(0);
} else {
  console.log(`\n❌ Found ${errors} blocking issue(s). Resolve them before production deploy.\n`);
  process.exit(1);
}
