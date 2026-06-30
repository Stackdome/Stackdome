import { Globe, Database, Zap, Box, type LucideIcon } from "lucide-react";
// Reuse the addon brand art for postgres/redis so the same software shows the
// same logo everywhere; mysql/mongo live in the shared brand folder.
import postgresUrl from "@/assets/addons/postgresql.svg";
import redisUrl from "@/assets/addons/redis.svg";
import mysqlUrl from "@/assets/brand/mysql.svg";
import mongoUrl from "@/assets/brand/mongodb.svg";

const LUCIDE: Record<string, LucideIcon> = { globe: Globe, database: Database, zap: Zap, box: Box };
const BRAND: Record<string, string> = { postgres: postgresUrl, redis: redisUrl, mysql: mysqlUrl, mongo: mongoUrl };

export function BlockGlyph({ icon, size = 18 }: { icon: string; size?: number }) {
  const brand = BRAND[icon];
  if (brand) {
    return (
      <img
        src={brand}
        alt=""
        aria-hidden
        style={{ width: size, height: size }}
        className="object-contain"
      />
    );
  }
  const Icon = LUCIDE[icon] ?? Box;
  return <Icon style={{ width: size, height: size }} />;
}
