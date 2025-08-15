#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const https = require('https');
const { execSync, spawn } = require('child_process');

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

// Binary paths
const binDir = __dirname;
const binaryName = platform === 'win32' ? 'openax.exe' : 'openax';
const binaryPath = path.join(binDir, binaryName);

// Check if binary exists, download if not
async function ensureBinary() {
  if (fs.existsSync(binaryPath)) {
    return; // Binary already exists
  }

  console.log(`Downloading OpenAx v${version} for ${platform} ${arch}...`);
  
  // Construct download URL
  const extension = platform === 'win32' ? '.zip' : '.tar.gz';
  const filename = `openax_${goPlatform}_${goArch}${extension}`;
  const downloadUrl = `https://github.com/imtanmoy/openax/releases/download/v${version}/${filename}`;
  const tempFile = path.join(binDir, filename);

  try {
    await downloadFile(downloadUrl, tempFile);
    await extractBinary(tempFile, binaryPath);
    fs.unlinkSync(tempFile); // Clean up
    console.log('✅ OpenAx downloaded successfully!');
  } catch (err) {
    console.error('❌ Failed to download OpenAx:', err.message);
    console.error('Please check: https://github.com/imtanmoy/openax/releases');
    process.exit(1);
  }
}

// Download file with redirect handling
function downloadFile(url, outputPath) {
  return new Promise((resolve, reject) => {
    function download(downloadUrl, followRedirects = true) {
      const file = fs.createWriteStream(outputPath);
      
      const request = https.get(downloadUrl, (response) => {
        // Handle redirects
        if (response.statusCode === 302 || response.statusCode === 301) {
          if (followRedirects && response.headers.location) {
            file.close();
            try {
              if (fs.existsSync(outputPath)) {
                fs.unlinkSync(outputPath);
              }
            } catch (e) {
              // Ignore cleanup errors
            }
            download(response.headers.location, false);
            return;
          }
        }
        
        if (response.statusCode !== 200) {
          file.close();
          reject(new Error(`HTTP ${response.statusCode}`));
          return;
        }

        response.pipe(file);
        
        file.on('finish', () => {
          file.close(() => resolve());
        });

        file.on('error', (err) => {
          file.close();
          reject(err);
        });
      });
      
      request.on('error', (err) => {
        reject(err);
      });
      
      request.setTimeout(30000, () => {
        request.destroy();
        reject(new Error('Download timeout'));
      });
    }
    
    download(url);
  });
}

// Extract binary from archive
async function extractBinary(archivePath, outputPath) {
  const dir = path.dirname(archivePath);
  
  if (platform === 'win32') {
    // Extract zip
    execSync(`cd "${dir}" && tar -xf "${path.basename(archivePath)}"`, { stdio: 'inherit' });
    fs.renameSync(path.join(dir, 'openax.exe'), outputPath);
  } else {
    // Extract tar.gz
    execSync(`cd "${dir}" && tar -xzf "${path.basename(archivePath)}"`, { stdio: 'inherit' });
    fs.renameSync(path.join(dir, 'openax'), outputPath);
    // Make executable
    fs.chmodSync(outputPath, '755');
  }
}

// Main execution
async function main() {
  try {
    await ensureBinary();
    
    // Execute the actual binary with all arguments
    const child = spawn(binaryPath, process.argv.slice(2), {
      stdio: 'inherit'
    });
    
    child.on('close', (code) => {
      process.exit(code);
    });
    
  } catch (err) {
    console.error('❌ Error:', err.message);
    process.exit(1);
  }
}

main();