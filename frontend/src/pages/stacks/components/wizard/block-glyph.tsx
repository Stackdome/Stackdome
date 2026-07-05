import { Globe, Database, Zap, Box, type LucideIcon } from "lucide-react";
// Reuse the addon brand art for postgres/redis so the same software shows the
// same logo everywhere; the rest live in the shared brand folder.
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

const LUCIDE: Record<string, LucideIcon> = { globe: Globe, database: Database, zap: Zap, box: Box };

// `light` renders in light mode (chip is light), `dark` renders in dark mode
// (chip is dark). Most logos are colorful enough to use one art for both; the
// near-mono ones (mariadb navy, influxdb navy, clickhouse yellow) need a
// contrasting variant so they stay legible on both chips.
const BRAND: Record<string, { light: string; dark: string }> = {
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
};

export function BlockGlyph({ icon, size = 18 }: { icon: string; size?: number }) {
  const brand = BRAND[icon];
  if (brand) {
    const dims = { width: size, height: size };
    return (
      <>
        <img src={brand.light} alt="" aria-hidden style={dims} className="object-contain dark:hidden" />
        <img src={brand.dark} alt="" aria-hidden style={dims} className="hidden object-contain dark:block" />
      </>
    );
  }
  const Icon = LUCIDE[icon] ?? Box;
  return <Icon style={{ width: size, height: size }} />;
}
