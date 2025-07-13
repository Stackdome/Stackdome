#!/usr/bin/env tsx

/**
 * Docker Compose to StackResource Conversion Checker
 *
 * This script uses our actual parser and converter modules to check if a given
 * Docker Compose file (local path or URL) can be successfully converted to a StackResource.
 *
 * Usage:
 *   pnpm run check:compose-conversion [options] <file-path-or-url>
 *
 * Options:
 *   --verbose, -v    Show detailed parsing and conversion information
 *   --show-output    Show the converted StackResource object
 *   --help, -h       Show this help message
 *
 * Examples:
 *   pnpm run check:compose-conversion ./__tests__/fixtures/simple-docker-compose.yml
 *   pnpm run check:compose-conversion --verbose https://example.com/docker-compose.yml
 *   pnpm run check:compose-conversion --show-output ./my-compose.yml
 *   pnpm run check:compose-conversion -v --show-output ./complex-compose.yml
 */

import { readFile } from 'fs/promises';
import { existsSync } from 'fs';
import { resolve, dirname } from 'path';
import { fileURLToPath } from 'url';
import { parseDockerCompose } from '@/lib/docker-compose-parser';
import { convertDockerComposeToStackData, getConversionSummary } from '@/lib/docker-compose-converter';

// Get __dirname equivalent in ES modules
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Colors for console output
const colors = {
  reset: '\x1b[0m',
  bright: '\x1b[1m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m',
} as const;

function colorize(text: string, color: keyof typeof colors): string {
  return `${colors[color]}${text}${colors.reset}`;
}

function printHeader(): void {
  console.log(colorize('\n🐳 Docker Compose to StackResource Conversion Checker', 'cyan'));
  console.log(colorize('='.repeat(55), 'cyan'));
}

function printUsage(): void {
  console.log(colorize('\nUsage:', 'bright'));
  console.log('  pnpm run check:compose-conversion [options] <file-path-or-url>');
  console.log(colorize('\nOptions:', 'bright'));
  console.log('  --verbose, -v    Show detailed parsing and conversion information');
  console.log('  --show-output    Show the converted StackResource object');
  console.log('  --help, -h       Show this help message');
  console.log(colorize('\nExamples:', 'bright'));
  console.log('  pnpm run check:compose-conversion ./__tests__/fixtures/simple-docker-compose.yml');
  console.log('  pnpm run check:compose-conversion --verbose https://example.com/docker-compose.yml');
  console.log('  pnpm run check:compose-conversion --show-output ./my-compose.yml');
  console.log('  pnpm run check:compose-conversion -v --show-output ./complex-compose.yml');
}

async function fetchFromUrl(url: string): Promise<string> {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`Failed to fetch ${url}: ${response.status} ${response.statusText}`);
  }
  return response.text();
}

async function readDockerComposeContent(input: string): Promise<string> {
  console.log(colorize(`📖 Reading Docker Compose content from: ${input}`, 'blue'));

  try {
    if (input.startsWith('http://') || input.startsWith('https://')) {
      return await fetchFromUrl(input);
    } else {
      // Handle relative paths - resolve from current working directory
      const filePath = resolve(input);

      if (!existsSync(filePath)) {
        throw new Error(`File not found: ${filePath}`);
      }

      return await readFile(filePath, 'utf-8');
    }
  } catch (error) {
    throw new Error(`Failed to read Docker Compose content: ${(error as Error).message}`);
  }
}

function printParseResult(parseResult: ReturnType<typeof parseDockerCompose>): void {
  console.log(colorize('\n📝 Parse Results:', 'bright'));

  if (parseResult.success) {
    console.log(colorize('✅ Parsing: SUCCESS', 'green'));

    const compose = parseResult.data!;
    const serviceCount = Object.keys(compose.services || {}).length;
    const volumeCount = Object.keys(compose.volumes || {}).length;
    const networkCount = Object.keys(compose.networks || {}).length;

    console.log(`   Services: ${serviceCount}`);
    console.log(`   Volumes: ${volumeCount}`);
    console.log(`   Networks: ${networkCount}`);

    if (compose.name) {
      console.log(`   Stack Name: ${compose.name}`);
    }
  } else {
    console.log(colorize('❌ Parsing: FAILED', 'red'));
    parseResult.errors?.forEach((error, index) => {
      console.log(colorize(`   Error ${index + 1}: ${error.message}`, 'red'));
      if (error.details) {
        console.log(colorize(`   Details: ${error.details}`, 'yellow'));
      }
    });
  }
}

function printConversionResult(conversionResult: ReturnType<typeof convertDockerComposeToStackData>): void {
  console.log(colorize('\n🔄 Conversion Results:', 'bright'));

  if (conversionResult.success) {
    console.log(colorize('✅ Conversion: SUCCESS', 'green'));

    const summary = getConversionSummary(conversionResult);
    console.log(`   Services Converted: ${summary.servicesConverted}`);
    console.log(`   Volumes Converted: ${summary.volumesConverted}`);
    console.log(`   Warnings: ${summary.warningCount}`);
    console.log(`   Errors: ${summary.errorCount}`);

    if (summary.requiresManualConfiguration) {
      console.log(colorize('   ⚠️  Manual configuration required', 'yellow'));
    }

    // Show generated stack name
    if (conversionResult.data?.name) {
      console.log(`   Generated Stack Name: ${conversionResult.data.name}`);
    }
  } else {
    console.log(colorize('❌ Conversion: FAILED', 'red'));
    conversionResult.errors?.forEach((error, index) => {
      console.log(colorize(`   Error ${index + 1}: ${error.message}`, 'red'));
      if (error.service) {
        console.log(colorize(`   Service: ${error.service}`, 'yellow'));
      }
      if (error.volume) {
        console.log(colorize(`   Volume: ${error.volume}`, 'yellow'));
      }
    });
  }

  // Show conversion errors even for successful conversions (partial successes)
  if (conversionResult.errors && conversionResult.errors.length > 0) {
    console.log(colorize('\n🚫 Conversion Errors:', 'red'));
    conversionResult.errors.forEach((error, index) => {
      console.log(colorize(`   ${index + 1}. ${error.message}`, 'red'));
      if (error.service) {
        console.log(colorize(`      Service: ${error.service}`, 'blue'));
      }
      if (error.volume) {
        console.log(colorize(`      Volume: ${error.volume}`, 'blue'));
      }
    });
  }
}

function printWarnings(conversionResult: ReturnType<typeof convertDockerComposeToStackData>): void {
  if (conversionResult.warnings && conversionResult.warnings.length > 0) {
    console.log(colorize('\n⚠️  Warnings:', 'yellow'));
    conversionResult.warnings.forEach((warning, index) => {
      console.log(colorize(`   ${index + 1}. ${warning.message}`, 'yellow'));
      if (warning.service) {
        console.log(colorize(`      Service: ${warning.service}`, 'blue'));
      }
      if (warning.volume) {
        console.log(colorize(`      Volume: ${warning.volume}`, 'blue'));
      }
      if (warning.dockerComposeField) {
        console.log(colorize(`      Field: ${warning.dockerComposeField}`, 'blue'));
      }
    });
  }
}

function printConversionSummary(
  parseResult: ReturnType<typeof parseDockerCompose>,
  conversionResult: ReturnType<typeof convertDockerComposeToStackData>
): void {
  console.log(colorize('\n📊 Summary:', 'bright'));

  const canConvert = parseResult.success && conversionResult.success;

  if (canConvert) {
    console.log(colorize('✅ This Docker Compose file CAN be converted to a StackResource!', 'green'));

    const summary = getConversionSummary(conversionResult);
    if (summary.warningCount > 0) {
      console.log(colorize(`   Note: ${summary.warningCount} warning(s) - review recommended`, 'yellow'));
    }
    if (summary.requiresManualConfiguration) {
      console.log(colorize('   Note: Manual configuration will be required for some features', 'yellow'));
    }
  } else {
    console.log(colorize('❌ This Docker Compose file CANNOT be converted to a StackResource', 'red'));

    if (!parseResult.success) {
      console.log(colorize('   Reason: Docker Compose file contains parsing errors', 'red'));
    } else if (!conversionResult.success) {
      console.log(colorize('   Reason: Conversion failed due to unsupported features or errors', 'red'));
    }
  }
}

function printConvertedObject(conversionResult: ReturnType<typeof convertDockerComposeToStackData>): void {
  if (conversionResult.success && conversionResult.data) {
    console.log(colorize('\n📋 Converted StackResource Object:', 'bright'));
    console.log(colorize('='.repeat(40), 'cyan'));

    try {
      const formattedOutput = JSON.stringify(conversionResult.data, null, 2);
      console.log(formattedOutput);
    } catch (error) {
      console.log(colorize('❌ Error formatting converted object:', 'red'));
      console.log((error as Error).message);
    }
  } else {
    console.log(colorize('\n❌ No converted object available (conversion failed)', 'red'));
  }
}

interface CliOptions {
  verbose: boolean;
  showOutput: boolean;
  help: boolean;
  input?: string;
}

function parseCliArgs(args: string[]): CliOptions {
  const options: CliOptions = {
    verbose: false,
    showOutput: false,
    help: false,
  };

  const remaining: string[] = [];

  for (let i = 0; i < args.length; i++) {
    const arg = args[i];

    switch (arg) {
      case '--verbose':
      case '-v':
        options.verbose = true;
        break;
      case '--show-output':
        options.showOutput = true;
        break;
      case '--help':
      case '-h':
        options.help = true;
        break;
      default:
        if (!arg.startsWith('-')) {
          remaining.push(arg);
        } else {
          console.log(colorize(`⚠️  Unknown option: ${arg}`, 'yellow'));
        }
        break;
    }
  }

  if (remaining.length > 0) {
    options.input = remaining[0];
  }

  return options;
}

async function main(): Promise<void> {
  const args = process.argv.slice(2);
  const options = parseCliArgs(args);

  printHeader();

  if (options.help || !options.input) {
    printUsage();
    process.exit(options.help ? 0 : 1);
  }

  try {
    // Step 1: Read the Docker Compose content
    const yamlContent = await readDockerComposeContent(options.input!);
    console.log(colorize('✅ Content read successfully', 'green'));

    // Step 2: Parse and validate the Docker Compose file
    const parseResult = parseDockerCompose(yamlContent);

    if (options.verbose) {
      printParseResult(parseResult);
    }

    if (!parseResult.success) {
      console.log(colorize('\n❌ Cannot proceed with conversion due to parsing errors', 'red'));
      process.exit(1);
    }

    // Step 3: Attempt conversion to StackResource
    const conversionResult = convertDockerComposeToStackData(parseResult.data!);

    if (options.verbose) {
      printConversionResult(conversionResult);

      // Step 4: Show warnings if any
      printWarnings(conversionResult);
    }

    // Step 5: Show converted object if requested
    if (options.showOutput) {
      printConvertedObject(conversionResult);
    }

    // Step 6: Print final summary (always shown)
    printConversionSummary(parseResult, conversionResult);

    // Exit with appropriate code
    process.exit(conversionResult.success ? 0 : 1);

  } catch (error) {
    console.log(colorize(`\n❌ Error: ${(error as Error).message}`, 'red'));
    process.exit(1);
  }
}

// Run the script
if (import.meta.url === `file://${process.argv[1]}`) {
  main().catch((error) => {
    console.error(colorize(`\n💥 Unexpected error: ${error.message}`, 'red'));
    process.exit(1);
  });
}
