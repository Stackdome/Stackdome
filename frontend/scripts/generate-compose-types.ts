import { compileFromFile } from 'json-schema-to-typescript';
import { writeFileSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

async function generateTypes() {
  try {
    const schemaPath = join(__dirname, '..', 'schemas', 'compose-spec.json');
    const outputPath = join(__dirname, '..', 'src', 'types', 'docker-compose-generated.ts');

    console.log('Generating Docker Compose types from schema...');

    const ts = await compileFromFile(schemaPath, {
      bannerComment: '/* Docker Compose Types - Generated from official schema */',
      style: {
        bracketSpacing: true,
        printWidth: 100,
        semi: true,
        singleQuote: true,
        tabWidth: 2,
        trailingComma: 'es5',
      },
      additionalProperties: false,
      enableConstEnums: true,
      format: true,
    });

    writeFileSync(outputPath, ts);
    console.log(`✅ Docker Compose types generated at: ${outputPath}`);
  } catch (error) {
    console.error('❌ Failed to generate Docker Compose types:', error);
    process.exit(1);
  }
}

generateTypes();
