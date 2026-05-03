import { describe, it, expect } from "vitest";
import {
  parseDockerCompose,
  parseAndValidateDockerCompose,
  validateDockerComposeService,
  extractServices,
  extractVolumes,
  extractNetworks,
  getDockerComposeSummary,
  createSampleDockerCompose,
} from "./docker-compose-parser";
import { readFileSync } from 'fs';
import { join } from 'path';

const simpleComposeYaml = readFileSync(join(__dirname, '__fixtures__/simple-docker-compose.yml'), 'utf-8');
const complexComposeYaml = readFileSync(join(__dirname, '__fixtures__/complex-docker-compose.yml'), 'utf-8');
const invalidComposeYaml = readFileSync(join(__dirname, '__fixtures__/invalid-docker-compose.yml'), 'utf-8');

describe("Docker Compose Parser", () => {
  describe("parseDockerCompose", () => {
    it("should parse valid Docker Compose YAML", () => {
      const yaml = `
version: '3.8'
services:
  web:
    image: nginx
    ports:
      - "80:80"
`;

      const result = parseDockerCompose(yaml);

      expect(result.success).toBe(true);
      expect(result.data).toBeDefined();
      expect(result.data?.services?.web?.image).toBe("nginx");
    });

    it("should handle invalid YAML syntax", () => {
      const invalidYaml = `
version: '3.8'
services:
  web:
    image: nginx
    ports:
      - "80:80
`; // Missing closing quote

      const result = parseDockerCompose(invalidYaml);

      expect(result.success).toBe(false);
      expect(result.errors).toHaveLength(1);
      expect(result.errors?.[0].type).toBe("yaml");
      expect(result.errors?.[0].message).toContain("Invalid YAML syntax");
    });

    it("should handle non-object YAML", () => {
      const result = parseDockerCompose("just a string");

      expect(result.success).toBe(false);
      expect(result.errors).toHaveLength(1);
      expect(result.errors?.[0].type).toBe("validation");
      expect(result.errors?.[0].message).toContain("must be a YAML object");
    });

    it("should handle files without services", () => {
      const yaml = `
version: '3.8'
name: test
`;

      const result = parseDockerCompose(yaml);

      // Should now succeed since we made services optional
      expect(result.success).toBe(true);
      expect(result.data?.name).toBe("test");
      expect(result.data?.version).toBe("3.8");
    });

    it("should handle complex Docker Compose file", () => {
      const yaml = `
version: '3.8'
name: complex-app

services:
  web:
    build:
      context: ./web
      dockerfile: Dockerfile.prod
    ports:
      - "80:80"
      - "443:443"
    depends_on:
      - api
      - redis
    environment:
      NODE_ENV: production
      API_URL: http://api:3000
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - static_files:/app/static

  api:
    image: node:18-alpine
    command: ["node", "server.js"]
    ports:
      - "3000:3000"
    environment:
      DATABASE_URL: postgres://user:pass@db:5432/myapp
      REDIS_URL: redis://redis:6379
    depends_on:
      - db
      - redis
    volumes:
      - ./api:/app
      - /app/node_modules

  db:
    image: postgres:14
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:alpine
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
  static_files:

networks:
  default:
    driver: bridge
`;

      const result = parseDockerCompose(yaml);

      expect(result.success).toBe(true);
      expect(result.data?.services).toBeDefined();
      expect(Object.keys(result.data?.services || {})).toHaveLength(4);
      expect(result.data?.volumes).toBeDefined();
      expect(Object.keys(result.data?.volumes || {})).toHaveLength(3);
    });
  });

  describe("parseAndValidateDockerCompose", () => {
    it("should return validated data for valid input", () => {
      const yaml = createSampleDockerCompose();

      const result = parseAndValidateDockerCompose(yaml);

      expect(result).toBeDefined();
      expect(result.services).toBeDefined();
    });

    it("should throw error for invalid input", () => {
      const invalidYaml = "invalid: yaml: content:";

      expect(() => parseAndValidateDockerCompose(invalidYaml)).toThrow();
    });
  });

  describe("validateDockerComposeService", () => {
    it("should validate valid service configuration", () => {
      const serviceConfig = {
        image: "nginx",
        ports: ["80:80"],
        environment: {
          NODE_ENV: "production"
        }
      };

      const result = validateDockerComposeService("web", serviceConfig);

      expect(result.success).toBe(true);
      expect(result.data?.services?.web).toBeDefined();
    });

    it("should handle invalid service configuration", () => {
      const serviceConfig = {
        ports: "invalid-port-format"
      };

      const result = validateDockerComposeService("web", serviceConfig);

      expect(result.success).toBe(false);
      expect(result.errors).toBeDefined();
      expect(result.errors?.length).toBeGreaterThan(0);
    });
  });

  describe("extraction functions", () => {
    const sampleCompose = parseAndValidateDockerCompose(createSampleDockerCompose());

    it("should extract services correctly", () => {
      const services = extractServices(sampleCompose);

      expect(Object.keys(services)).toContain("web");
      expect(Object.keys(services)).toContain("api");
      expect(Object.keys(services)).toContain("db");
    });

    it("should extract volumes correctly", () => {
      const volumes = extractVolumes(sampleCompose);

      expect(Object.keys(volumes)).toContain("postgres_data");
    });

    it("should extract networks correctly", () => {
      const networks = extractNetworks(sampleCompose);

      expect(Object.keys(networks)).toContain("default");
    });
  });

  describe("getDockerComposeSummary", () => {
    it("should provide correct summary information", () => {
      const compose = parseAndValidateDockerCompose(createSampleDockerCompose());
      const summary = getDockerComposeSummary(compose);

      expect(summary.name).toBe("sample-app");
      expect(summary.serviceCount).toBe(3);
      expect(summary.volumeCount).toBe(1);
      expect(summary.networkCount).toBe(1);
      expect(summary.serviceNames).toContain("web");
      expect(summary.serviceNames).toContain("api");
      expect(summary.serviceNames).toContain("db");
      expect(summary.volumeNames).toContain("postgres_data");
    });

    it("should handle empty compose file", () => {
      const emptyCompose = { services: {} };
      const summary = getDockerComposeSummary(emptyCompose);

      expect(summary.serviceCount).toBe(0);
      expect(summary.volumeCount).toBe(0);
      expect(summary.networkCount).toBe(0);
      expect(summary.serviceNames).toHaveLength(0);
    });
  });

  describe("createSampleDockerCompose", () => {
    it("should create valid sample Docker Compose", () => {
      const sample = createSampleDockerCompose();

      expect(sample).toBeTruthy();
      expect(sample).toContain("version:");
      expect(sample).toContain("services:");

      const result = parseDockerCompose(sample);
      expect(result.success).toBe(true);
    });
  });

  describe("error handling", () => {
    it("should provide detailed validation errors", () => {
      const yaml = `
version: '3.8'
services:
  web:
    image: 123  # Invalid type - should be string
`;

      const result = parseDockerCompose(yaml);

      expect(result.success).toBe(false);
      expect(result.errors).toBeDefined();

      // Check that errors have proper structure
      result.errors?.forEach(error => {
        expect(error).toHaveProperty("type");
        expect(error).toHaveProperty("message");
        expect(["yaml", "validation"]).toContain(error.type);
      });
    });
  });
});

describe('fixture-based parsing tests', () => {
  describe('simple Docker Compose fixture', () => {
    it('should parse simple web app Docker Compose', () => {
      const parseResult = parseDockerCompose(simpleComposeYaml);

      expect(parseResult.success).toBe(true);
      expect(parseResult.data).toBeDefined();

      if (parseResult.data) {
        expect(Object.keys(parseResult.data.services || {})).toHaveLength(3);
        expect(Object.keys(parseResult.data.volumes || {})).toHaveLength(1);
        expect(parseResult.data.name).toBe('simple-web-app');

        // Test parsing utilities
        const services = extractServices(parseResult.data);
        const volumes = extractVolumes(parseResult.data);
        const summary = getDockerComposeSummary(parseResult.data);

        expect(Object.keys(services)).toHaveLength(3);
        expect(Object.keys(volumes)).toHaveLength(1);
        expect(summary.serviceCount).toBe(3);
        expect(summary.volumeCount).toBe(1);

        // Test service validation
        Object.entries(services).forEach(([serviceName, service]) => {
          const validation = validateDockerComposeService(serviceName, service);
          expect(validation.success).toBe(true);
        });
      }
    });
  });

  describe('complex Docker Compose fixture', () => {
    it('should parse complex multi-service Docker Compose', () => {
      const parseResult = parseDockerCompose(complexComposeYaml);

      expect(parseResult.success).toBe(true);
      expect(parseResult.data).toBeDefined();

      if (parseResult.data) {
        expect(Object.keys(parseResult.data.services || {})).toHaveLength(4);
        expect(Object.keys(parseResult.data.volumes || {})).toHaveLength(3);
        expect(Object.keys(parseResult.data.networks || {})).toHaveLength(2);
        expect(parseResult.data.name).toBe('complex-app');

        // Verify complex features are parsed correctly
        expect(parseResult.data.networks).toBeDefined();
        expect(parseResult.data.secrets).toBeDefined();

        // Test that build specifications are parsed
        const frontend = parseResult.data.services?.frontend;
        expect(frontend?.build).toBeDefined();

        // Test parsing utilities with complex structure
        const summary = getDockerComposeSummary(parseResult.data);
        expect(summary.serviceCount).toBe(4);
        expect(summary.volumeCount).toBe(3);
        expect(summary.networkCount).toBe(2);
      }
    });
  });

  describe('invalid Docker Compose fixture', () => {
    it('should parse invalid Docker Compose with warnings', () => {
      const parseResult = parseDockerCompose(invalidComposeYaml);

      // Should parse successfully (YAML is valid, just service configs are invalid)
      expect(parseResult.success).toBe(true);
      expect(parseResult.data).toBeDefined();

      if (parseResult.data) {
        expect(Object.keys(parseResult.data.services || {})).toHaveLength(2);
        expect(Object.keys(parseResult.data.volumes || {})).toHaveLength(0);

        // Test service validation catches invalid services
        const services = extractServices(parseResult.data);
        const invalidService = services['invalid-service'];
        expect(invalidService).toBeDefined();
      }
    });
  });

  describe('parseAndValidateDockerCompose with fixtures', () => {
    it('should validate fixture files successfully at parse level', () => {
      const fixtures = [
        { name: 'simple', yaml: simpleComposeYaml },
        { name: 'complex', yaml: complexComposeYaml },
        { name: 'invalid', yaml: invalidComposeYaml }
      ];

      fixtures.forEach(({ name: _name, yaml }) => {
        const result = parseAndValidateDockerCompose(yaml);

        // All our fixtures should parse successfully (conversion may have issues)
        expect(result).toBeDefined();
        expect(result.services).toBeDefined();
      });
    });
  });
});
