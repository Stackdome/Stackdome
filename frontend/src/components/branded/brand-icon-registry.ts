// Central brand-icon registry: slug → themed SVG art. One place to grow the
// icon set — add the asset under assets/brand (selfh.st/icons is the usual
// source for missing ones), register it here, and every consumer (canvas
// nodes, wizard blocks, future surfaces) picks it up.
//
// `light` renders in light mode, `dark` in dark mode. Most logos are colorful
// enough to share one art; near-mono ones (mariadb navy, influxdb navy,
// clickhouse yellow, minio's black wordmark) need a contrasting variant so
// they stay legible on both surfaces.
import postgresUrl from "@/assets/addons/postgresql.svg";
import redisUrl from "@/assets/addons/redis.svg";
import mysqlUrl from "@/assets/brand/mysql.svg";
import mongoUrl from "@/assets/brand/mongodb.svg";
import mariadbUrl from "@/assets/brand/mariadb.svg";
import mariadbLightUrl from "@/assets/brand/mariadb-light.svg";
import mssqlUrl from "@/assets/brand/mssql.svg";
import elasticsearchUrl from "@/assets/brand/elasticsearch.svg";
import couchdbUrl from "@/assets/brand/couchdb.svg";
import influxdbUrl from "@/assets/brand/influxdb.svg";
import influxdbLightUrl from "@/assets/brand/influxdb-light.svg";
import clickhouseUrl from "@/assets/brand/clickhouse.svg";
import clickhouseDarkUrl from "@/assets/brand/clickhouse-dark.svg";
import tooljetUrl from "@/assets/brand/tooljet.svg";
import grafanaUrl from "@/assets/brand/grafana.svg";
import opentelemetryUrl from "@/assets/brand/opentelemetry.svg";
import minioUrl from "@/assets/brand/minio.svg";
import minioLightUrl from "@/assets/brand/minio-light.svg";
import postgrestUrl from "@/assets/brand/postgrest.png";

export const BRAND_ICONS: Record<string, { light: string; dark: string }> = {
  postgres: { light: postgresUrl, dark: postgresUrl },
  redis: { light: redisUrl, dark: redisUrl },
  mysql: { light: mysqlUrl, dark: mysqlUrl },
  mongo: { light: mongoUrl, dark: mongoUrl },
  mariadb: { light: mariadbUrl, dark: mariadbLightUrl },
  mssql: { light: mssqlUrl, dark: mssqlUrl },
  elasticsearch: { light: elasticsearchUrl, dark: elasticsearchUrl },
  couchdb: { light: couchdbUrl, dark: couchdbUrl },
  influxdb: { light: influxdbUrl, dark: influxdbLightUrl },
  clickhouse: { light: clickhouseDarkUrl, dark: clickhouseUrl },
  tooljet: { light: tooljetUrl, dark: tooljetUrl },
  grafana: { light: grafanaUrl, dark: grafanaUrl },
  opentelemetry: { light: opentelemetryUrl, dark: opentelemetryUrl },
  minio: { light: minioUrl, dark: minioLightUrl },
  // PNG (GitHub org mark) until a clean SVG exists; light bg is baked in.
  postgrest: { light: postgrestUrl, dark: postgrestUrl },
};

export function hasBrandIcon(slug: string | undefined): slug is string {
  return !!slug && slug in BRAND_ICONS;
}
