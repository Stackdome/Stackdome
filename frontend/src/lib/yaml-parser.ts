import { load } from "js-yaml";

type StackCompose = Record<string, any>;

// Sample stack configuration for demo purposes
const SAMPLE_STACK = `
nginx:
  image: nginx
  dependsOn:
    - backend
    - frontend
  volumeMounts:
    code/nginx/default.dev.conf: /etc/nginx/conf.d/default.conf
  ports:
  - number: 80
    exposeToPublic: true
    isHttp: true

db:
  image: postgres:14-alpine
  volumeMounts:
    postgresdata: /var/lib/postgresql/data
  environmentVariables:
    POSTGRES_PASSWORD: infisical
    POSTGRES_USER: infisical
    POSTGRES_DB: infisical
  ports:
  - number: 5432
    exposeToPublic: false
    isHttp: false

redis:
  image: redis
  environmentVariables:
    ALLOW_EMPTY_PASSWORD: yes
  volumeMounts:
    redisdata: /data
  ports:
  - number: 6379
    exposeToPublic: false
    isHttp: false

backend:
  imageRegistry: k8s.orb.local:5000/infisical-backend
  build:
    sourceVolume: code
    buildContext: /backend
    dockerFilePath: /backend/Dockerfile.dev
  command: ["/bin/sh", "-c"]
  args:
  - |
    npm run migration:latest && npm run dev:docker
  dependsOn:
    - db
    - redis
  envFiles:
  - .env
  ports:
  - number: 4000
    exposeToPublic: true
    isHttp: true
  environmentVariables:
    NODE_ENV: development
    DB_CONNECTION_URI: postgres://infisical:infisical@db/infisical?sslmode=disable
    TELEMETRY_ENABLED: "false"
  volumeMounts:
    code/backend/src: /app/src

frontend:
  imageRegistry: k8s.orb.local:5000/infisical-frontend
  dependsOn:
    - backend
    - redis
    - db
  build:
    sourceVolume: code
    buildContext: /frontend
    dockerFilePath: /frontend/Dockerfile.dev
  ports:
  - number: 3000
    exposeToPublic: true
    isHttp: true
  volumeMounts:
    code/frontend/src: /app/src/
    code/frontend/public: /app/public
  envFiles:
  - .env
  environmentVariables:
    NEXT_PUBLIC_ENV: development

volumes:
  postgresdata:
    size: 2Gi
  redisdata:
    size: 2Gi
  code:
    size: 2Gi
    source:
      localDir:
        sync: true
        path: .
`;

export function parseYaml(yamlContent: string): StackCompose {
  try {
    const parsed = load(yamlContent) as Record<string, any>;
    
    // Basic validation
    if (typeof parsed !== 'object' || parsed === null) {
      throw new Error("Invalid YAML: Expected an object");
    }
    
    // Convert the parsed YAML to our StackCompose type
    const stackCompose: StackCompose = { ...parsed };
    
    // Remove any non-object properties except for 'volumes'
    Object.keys(stackCompose).forEach(key => {
      if (key !== 'volumes' && (typeof stackCompose[key] !== 'object' || stackCompose[key] === null)) {
        delete stackCompose[key];
      }
    });
    
    return stackCompose;
  } catch (error) {
    console.error("Failed to parse YAML:", error);
    throw new Error(`Invalid YAML format: ${(error as Error).message}`);
  }
}

export function getSampleStackYaml(): string {
  return SAMPLE_STACK.trim();
}
