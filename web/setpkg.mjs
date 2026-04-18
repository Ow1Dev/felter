import fs from "fs"; const p=JSON.parse(fs.readFileSync("package.json","utf8")); p.packageManager="bun@1.3.11"; fs.writeFileSync("package.json", JSON.stringify(p,null,2)+"\n");
