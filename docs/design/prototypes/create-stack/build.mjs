/**
 * Inlines the two product typefaces into prototype.source.html and writes
 * prototype.html, which opens standalone in any browser.
 *
 *   node docs/design/prototypes/create-stack/build.mjs
 *
 * The fonts are read from the embedded web build so the prototype is set in
 * exactly the faces the product ships.
 */
import { readFileSync, writeFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const here = dirname(fileURLToPath(import.meta.url));
const root = resolve(here, "../../../..");
const b64 = (p) => `data:font/woff2;base64,${readFileSync(p).toString("base64")}`;

let html = readFileSync(resolve(here, "prototype.source.html"), "utf8")
  .replace("__GEIST__", () => b64(resolve(root, "pkg/web/dist/assets/geist-latin-wght-normal-BgDaEnEv.woff2")))
  .replace("__MONO__", () => b64(resolve(root, "pkg/web/dist/assets/jetbrains-mono-latin-wght-normal-B9CIFXIH.woff2")));

const left = html.match(/__[A-Z0-9]+__/g);
if (left) throw new Error("unsubstituted placeholders: " + [...new Set(left)].join(", "));

writeFileSync(resolve(here, "prototype.html"), html);
console.log(`wrote prototype.html (${(html.length / 1024).toFixed(0)} KB)`);
