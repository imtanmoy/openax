#!/usr/bin/env node

const { execSync } = require('child_process');
const path = require('path');

const binPath = path.join(__dirname, '..', 'bin', process.platform === 'win32' ? 'openax.exe' : 'openax');

try {
  console.log('Testing OpenAx installation...');
  
  // Test version command
  const versionOutput = execSync(`"${binPath}" --version`, { encoding: 'utf8' });
  console.log('✅ Version check:', versionOutput.trim());
  
  // Test help command
  const helpOutput = execSync(`"${binPath}" --help`, { encoding: 'utf8' });
  if (helpOutput.includes('OpenAx') || helpOutput.includes('openax')) {
    console.log('✅ Help command works');
  } else {
    console.log('⚠️  Help command output unexpected');
  }
  
  console.log('✅ OpenAx is working correctly!');
  
} catch (err) {
  console.error('❌ Test failed:', err.message);
  process.exit(1);
}