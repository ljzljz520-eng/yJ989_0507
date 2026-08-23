import { mkdir, readFile, writeFile } from "node:fs/promises";

const html = await readFile("index.html", "utf8");
await mkdir("dist", { recursive: true });
await writeFile("dist/index.html", html.replace("src/main.js", "main.js"));
await writeFile("dist/main.js", await readFile("src/main.js", "utf8"));
console.log("activity registration web built");
