import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

test("the built explorer carries its shell and canonical data", async () => {
  const html = await readFile(new URL("../dist/index.html", import.meta.url), "utf8");
  assert.match(html, /Personality Palette \/\/ Agent Compose/);
  assert.match(html, /type="module"/);

  const raw = await readFile(new URL("../dist/palette.json", import.meta.url), "utf8");
  const palette = JSON.parse(raw);
  assert.equal(palette.version, 2);
  assert.ok(Array.isArray(palette.personalities) && palette.personalities.length > 0);
  assert.ok(Array.isArray(palette.expressions) && palette.expressions.length > 0);
  assert.ok(Array.isArray(palette.roles) && palette.roles.length > 0);
  assert.ok(palette.personalities.every((personality) => (
    typeof personality.motif === "string"
    && typeof personality.emblem?.name === "string"
    && typeof personality.form?.silhouette === "string"
    && typeof personality.sound_mark?.timbre === "string"
  )));
  assert.ok(palette.roles.every((role) => /^#[0-9a-f]{6}$/i.test(role.color)));
});
