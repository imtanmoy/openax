#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const https = require('https');
const { execSync } = require('child_process');

// Get package info
const packageJson = require('../package.json');
const version = packageJson.version;

// Determine platform and architecture
const platform = process.platform;
const arch = process.arch;

// Map Node.js platform/arch to Go build names
const platformMap = {
  'darwin': 'Darwin',
  'linux': 'Linux',
  'win32': 'Windows'
};

const archMap = {
  'x64': 'x86_64',
  'arm64': 'arm64'
};

const goPlatform = platformMap[platform];
const goArch = archMap[arch];

if (!goPlatform || !goArch) {
  console.error(`Unsupported platform: ${platform} ${arch}`);
  process.exit(1);
}

// Construct download URL
const extension = platform === 'win32' ? '.zip' : '.tar.gz';
const filename = `openax_${goPlatform}_${goArch}${extension}`;
const downloadUrl = `https://github.com/imtanmoy/openax/releases/download/v${version}/${filename}`;

console.log(`Installing OpenAx v${version} for ${platform} ${arch}...`);
console.log(`Downloading from: ${downloadUrl}`);

// Download and extract binary
const tempFile = path.join(__dirname, '..', filename);
const binDir = path.join(__dirname, '..', 'bin');
const binPath = path.join(binDir, platform === 'win32' ? 'openax.exe' : 'openax');

// Ensure bin directory exists
if (!fs.existsSync(binDir)) {
  fs.mkdirSync(binDir, { recursive: true });
}

// Download file
const file = fs.createWriteStream(tempFile);
https.get(downloadUrl, (response) => {
  if (response.statusCode !== 200) {
    console.error(`Failed to download: HTTP ${response.statusCode}`);
    console.error('This usually means the release for your platform is not available yet.');
    console.error('Please check: https://github.com/imtanmoy/openax/releases');
    process.exit(1);
  }

  response.pipe(file);
  
  file.on('finish', () => {
    file.close(() => {
      try {
        // Extract the binary
        if (platform === 'win32') {
          // For Windows, we need to unzip
          execSync(`cd "${path.dirname(tempFile)}" && tar -xf "${filename}"`, { stdio: 'inherit' });
          // Move binary to bin directory
          fs.renameSync(path.join(path.dirname(tempFile), 'openax.exe'), binPath);
        } else {
          // For Unix systems, extract tar.gz
          execSync(`cd "${path.dirname(tempFile)}" && tar -xzf "${filename}"`, { stdio: 'inherit' });
          // Move binary to bin directory
          fs.renameSync(path.join(path.dirname(tempFile), 'openax'), binPath);
        }

        // Make binary executable (Unix systems)
        if (platform !== 'win32') {
          fs.chmodSync(binPath, '755');
        }

        // Clean up temp file
        fs.unlinkSync(tempFile);

        console.log('✅ OpenAx installed successfully!');
        console.log(`Binary location: ${binPath}`);
        
        // Test the binary
        try {
          const output = execSync(`"${binPath}" --version`, { encoding: 'utf8' });
          console.log(`✅ ${output.trim()}`);
        } catch (err) {
          console.log('⚠️  Binary installed but version check failed');
        }

      } catch (err) {
        console.error('❌ Failed to extract binary:', err.message);
        process.exit(1);
      }
    });
  });

  file.on('error', (err) => {
    console.error('❌ Download failed:', err.message);
    process.exit(1);
  });
});