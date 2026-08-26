import "./styles.css";

type CanonicalPersonality = {
  name: string;
  color: string;
  motif: string;
  emblem: {
    names: string[];
    emoji: string;
  };
  geometry: string;
  body: {
    archetype: string;
    attachment: string;
  };
  soundMark: {
    timbre: string;
    contour: string;
    pulse: string;
  };
};

type CanonicalRole = {
  name: string;
  personalities: string[];
  color: string;
};

type PaletteDocument = {
  version: 2;
  expressions: string[];
  personalities: CanonicalPersonality[];
  roles: CanonicalRole[];
};

type Presentation = {
  name: string;
  colorName: string;
  association: string;
};

type SortMode = "spectrum" | "alphabetical";
type Theme = "day" | "night";

const presentation: Presentation[] = [
  { name: "decisive", colorName: "rose red", association: "committed cut" },
  { name: "warm", colorName: "coral", association: "open welcome" },
  { name: "outward", colorName: "brass yellow", association: "outside check" },
  { name: "tenacious", colorName: "weathered olive", association: "durable grip" },
  { name: "grounded", colorName: "sage", association: "steady growth" },
  { name: "protective", colorName: "emerald", association: "trusted shelter" },
  { name: "empirical", colorName: "aquamarine", association: "measured result" },
  { name: "immersed", colorName: "deep azure", association: "inside view" },
  { name: "imaginative", colorName: "violet", association: "visible possibility" },
  { name: "playful", colorName: "magenta", association: "electric delight" },
];

const appElement = document.querySelector<HTMLDivElement>("#app");
if (appElement === null) {
  throw new Error("palette app root is missing");
}
const app: HTMLDivElement = appElement;

const state: {
  role: string;
  sort: SortMode;
  theme: Theme;
  copied: string;
} = {
  role: "all",
  sort: "spectrum",
  theme: "day",
  copied: "",
};

const hexPattern = /^#[0-9a-f]{6}$/i;
const semanticTokenPattern = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function readStringList(
  record: Record<string, unknown>,
  key: string,
  label: string,
): string[] {
  const value = record[key];
  if (!Array.isArray(value) || value.length === 0) {
    throw new Error(`${label} needs ${key}`);
  }
  return value.map((entry) => {
    if (typeof entry !== "string" || !semanticTokenPattern.test(entry)) {
      throw new Error(`${label} ${key} needs lowercase semantic tokens`);
    }
    return entry;
  });
}

function readString(
  record: Record<string, unknown>,
  key: string,
  label: string,
  semantic = true,
): string {
  const value = record[key];
  if (
    typeof value !== "string"
    || value.length === 0
    || (semantic && !semanticTokenPattern.test(value))
  ) {
    throw new Error(`${label} needs ${key}`);
  }
  return value;
}

function parsePalette(value: unknown): PaletteDocument {
  if (!isRecord(value) || value.version !== 2) {
    throw new Error("palette data needs schema version 2");
  }
  if (
    !Array.isArray(value.expressions)
    || !value.expressions.every(
      (expression): expression is string => (
        typeof expression === "string" && semanticTokenPattern.test(expression)
      ),
    )
    || new Set(value.expressions).size !== value.expressions.length
    || !Array.isArray(value.personalities)
    || !Array.isArray(value.roles)
  ) {
    throw new Error("palette data needs unique expressions, personalities, and roles");
  }

  const personalities = value.personalities.map((entry, index): CanonicalPersonality => {
    if (
      !isRecord(entry)
      || typeof entry.name !== "string"
      || typeof entry.color !== "string"
      || !hexPattern.test(entry.color)
      || !isRecord(entry.emblem)
      || !isRecord(entry.body)
      || !isRecord(entry.sound_mark)
    ) {
      throw new Error(`palette personality ${index} is invalid`);
    }
    const label = `palette personality ${index}`;
    return {
      name: readString(entry, "name", label),
      color: entry.color.toLowerCase(),
      motif: readString(entry, "motif", label),
      emblem: {
        names: readStringList(entry.emblem, "names", `${label} emblem`),
        emoji: readString(entry.emblem, "emoji", `${label} emblem`, false),
      },
      geometry: readString(entry, "geometry", label),
      body: {
        archetype: readString(entry.body, "archetype", `${label} body`, false),
        attachment: readString(entry.body, "attachment", `${label} body`, false),
      },
      soundMark: {
        timbre: readString(entry.sound_mark, "timbre", `${label} sound mark`),
        contour: readString(entry.sound_mark, "contour", `${label} sound mark`),
        pulse: readString(entry.sound_mark, "pulse", `${label} sound mark`),
      },
    };
  });
  const personalityNames = new Set(personalities.map(({ name }) => name));
  if (personalityNames.size !== personalities.length) {
    throw new Error("palette personality names must be unique");
  }

  const roles = value.roles.map((entry, index): CanonicalRole => {
    if (
      !isRecord(entry)
      || typeof entry.name !== "string"
      || typeof entry.color !== "string"
      || !hexPattern.test(entry.color)
      || !Array.isArray(entry.personalities)
      || !entry.personalities.every(
        (name): name is string => typeof name === "string" && personalityNames.has(name),
      )
    ) {
      throw new Error(`palette role ${index} is invalid`);
    }
    return {
      name: entry.name,
      personalities: [...entry.personalities],
      color: entry.color.toLowerCase(),
    };
  });
  if (new Set(roles.map(({ name }) => name)).size !== roles.length) {
    throw new Error("palette role names must be unique");
  }

  return { version: 2, expressions: [...value.expressions], personalities, roles };
}

function escapeHTML(value: string): string {
  return value.replace(
    /[&<>"']/g,
    (character) => ({
      "&": "&amp;",
      "<": "&lt;",
      ">": "&gt;",
      '"': "&quot;",
      "'": "&#039;",
    })[character] ?? character,
  );
}

function roleLabel(value: string): string {
  return value.replaceAll("-", " ");
}

function ensurePresentationMatches(data: PaletteDocument): void {
  const canonical = new Set(data.personalities.map(({ name }) => name));
  const presented = new Set(presentation.map(({ name }) => name));
  const missing = [...canonical].filter((name) => !presented.has(name));
  const foreign = [...presented].filter((name) => !canonical.has(name));
  if (missing.length > 0 || foreign.length > 0) {
    throw new Error(
      `presentation metadata drifted, missing [${missing.join(", ")}], foreign [${foreign.join(", ")}]`,
    );
  }
}

function renderColorSentence(
  personalityByName: ReadonlyMap<string, CanonicalPersonality>,
): string {
  return presentation.map((item, index) => {
    const canonical = personalityByName.get(item.name);
    if (canonical === undefined) {
      throw new Error(`missing canonical personality ${item.name}`);
    }
    return `<span style="--personality-color:${canonical.color};--delay:${index * 28}ms">${escapeHTML(item.name)}</span>`;
  }).join("");
}

function renderRolePicker(roles: CanonicalRole[]): string {
  const buttons = [
    `<button type="button" class="${state.role === "all" ? "active" : ""}" aria-pressed="${state.role === "all"}" data-action="role" data-role="all">all</button>`,
    ...roles.map((role) => `
      <button
        type="button"
        class="${state.role === role.name ? "active" : ""}"
        aria-pressed="${state.role === role.name}"
        data-action="role"
        data-role="${escapeHTML(role.name)}"
      >
        <span class="role-dot" style="--role-color:${role.color}" aria-hidden="true"></span>
        ${escapeHTML(roleLabel(role.name))}
      </button>
    `),
  ];
  return buttons.join("");
}

function renderRoleMeld(
  role: CanonicalRole | undefined,
  personalityByName: ReadonlyMap<string, CanonicalPersonality>,
): string {
  if (role === undefined) {
    return "";
  }
  const componentColors = role.personalities.map((name) => {
    const personality = personalityByName.get(name);
    if (personality === undefined) {
      throw new Error(`role ${role.name} names missing personality ${name}`);
    }
    return `
      <span
        style="--component-color:${personality.color}"
        title="${escapeHTML(name)} ${personality.color}"
      ></span>
    `;
  }).join("");
  const copyLabel = state.copied === `${role.name} meld` ? "copied" : "copy";

  return `
    <div class="role-meld" style="--role-color:${role.color}">
      <div class="role-meld-copy">
        <p class="section-index">Role meld</p>
        <h3>${escapeHTML(roleLabel(role.name))}</h3>
        <p>${role.personalities.map(escapeHTML).join(" + ")}</p>
      </div>
      <div class="meld-equation" aria-label="${escapeHTML(roleLabel(role.name))} melded color ${role.color}">
        <div class="component-colors">${componentColors}</div>
        <span class="equals" aria-hidden="true">→</span>
        <button
          type="button"
          class="meld-result"
          data-action="copy"
          data-copy="${role.color}"
          data-key="${escapeHTML(`${role.name} meld`)}"
          aria-label="Copy ${role.color}, the melded color for ${escapeHTML(roleLabel(role.name))}"
        >
          <span class="meld-swatch"></span>
          <span>
            <small>Melded favorite</small>
            <strong>${role.color}</strong>
          </span>
          <em>${copyLabel}</em>
        </button>
      </div>
    </div>
  `;
}

function renderCards(
  visible: CanonicalPersonality[],
  presentationByName: ReadonlyMap<string, Presentation>,
): string {
  return visible.map((personality, index) => {
    const details = presentationByName.get(personality.name);
    if (details === undefined) {
      throw new Error(`missing presentation for ${personality.name}`);
    }
    const spectrum = presentation.findIndex(({ name }) => name === personality.name) + 1;
    const copyLabel = state.copied === personality.name ? "copied" : "copy";
    return `
      <article
        class="color-card"
        style="--personality-color:${personality.color};--card-index:${index}"
      >
        <div class="swatch">
          <span class="swatch-no">${String(spectrum).padStart(2, "0")}</span>
          <span
            class="personality-emblem"
            aria-label="${escapeHTML(personality.emblem.names[0] ?? personality.name)}"
            title="${escapeHTML(personality.emblem.names.join(" / "))}"
          >${escapeHTML(personality.emblem.emoji)}</span>
          <div class="contrast-pair" aria-label="Light and dark contrast sample">
            <span>Aa</span><span>Aa</span>
          </div>
        </div>
        <div class="card-copy">
          <p class="color-name">${escapeHTML(details.colorName)}</p>
          <h3>${escapeHTML(personality.name)}</h3>
          <p class="association">${escapeHTML(details.association)}</p>
          <dl class="identity-primitives">
            <div><dt>motif</dt><dd>${escapeHTML(personality.motif)}</dd></div>
            <div><dt>emblem</dt><dd>${escapeHTML(personality.emblem.names.join(" / "))}</dd></div>
            <div><dt>geometry</dt><dd>${escapeHTML(personality.geometry)}</dd></div>
            <div><dt>sound</dt><dd>${escapeHTML(personality.soundMark.timbre)}</dd></div>
          </dl>
          <button
            type="button"
            class="hex"
            data-action="copy"
            data-copy="${personality.color}"
            data-key="${escapeHTML(personality.name)}"
            aria-label="Copy ${personality.color}, the color for ${escapeHTML(personality.name)}"
          >
            <span>${personality.color}</span><span aria-hidden="true">${copyLabel}</span>
          </button>
        </div>
      </article>
    `;
  }).join("");
}

function render(data: PaletteDocument): void {
  const personalityByName = new Map(data.personalities.map((item) => [item.name, item]));
  const presentationByName = new Map(presentation.map((item) => [item.name, item]));
  const selectedRole = data.roles.find(({ name }) => name === state.role);
  const selectedNames = selectedRole === undefined
    ? new Set(data.personalities.map(({ name }) => name))
    : new Set(selectedRole.personalities);
  const visible = presentation
    .map(({ name }) => personalityByName.get(name))
    .filter((item): item is CanonicalPersonality => item !== undefined && selectedNames.has(item.name))
    .sort((left, right) => state.sort === "alphabetical"
      ? left.name.localeCompare(right.name)
      : presentation.findIndex(({ name }) => name === left.name)
        - presentation.findIndex(({ name }) => name === right.name));
  const heading = selectedRole === undefined ? "The full company" : roleLabel(selectedRole.name);
  const toolbarCopy = selectedRole === undefined
    ? "Every color is distinct in perceptual space."
    : `Every personality shown is active for the ${roleLabel(selectedRole.name)} seat.`;

  app.innerHTML = `
    <main class="site-shell" data-theme="${state.theme}">
      <div class="ambient ambient-one"></div>
      <div class="ambient ambient-two"></div>
      <header class="masthead">
        <a class="wordmark" href="#top" aria-label="Personality palette home">
          <span class="wordmark-mark" aria-hidden="true"><i></i><i></i><i></i><i></i></span>
          <span>agent compose</span>
        </a>
        <div class="theme-switch" aria-label="Preview background">
          <button type="button" aria-pressed="${state.theme === "day"}" data-action="theme" data-theme="day">Day</button>
          <button type="button" aria-pressed="${state.theme === "night"}" data-action="theme" data-theme="night">Night</button>
        </div>
      </header>

      <section class="hero" id="top">
        <div class="hero-copy">
          <p class="eyebrow">Canonical personality palette</p>
          <h1>Sixteen personalities.<span>One shared language of color.</span></h1>
          <p class="lede">
            A visual map for how each agent personality should feel before it
            says a word. Choose a role to see its complete personality meld.
          </p>
        </div>
        <div class="color-sentence" aria-label="All personality colors">
          ${renderColorSentence(personalityByName)}
        </div>
      </section>

      <section class="explorer" aria-labelledby="explorer-heading">
        <div class="explorer-heading">
          <div>
            <p class="section-index">01 / explore</p>
            <h2 id="explorer-heading">${escapeHTML(heading)}</h2>
          </div>
          <p class="result-count"><strong>${visible.length}</strong> personalities</p>
        </div>

        <div class="role-picker" aria-label="Filter by role">
          ${renderRolePicker(data.roles)}
        </div>
        ${renderRoleMeld(selectedRole, personalityByName)}

        <div class="toolbar">
          <p>${escapeHTML(toolbarCopy)}</p>
          <label>
            <span>Arrange</span>
            <select aria-label="Arrange personalities">
              <option value="spectrum" ${state.sort === "spectrum" ? "selected" : ""}>by spectrum</option>
              <option value="alphabetical" ${state.sort === "alphabetical" ? "selected" : ""}>alphabetically</option>
            </select>
          </label>
        </div>

        <div class="palette-grid">${renderCards(visible, presentationByName)}</div>
      </section>

      <footer>
        <p>Color is expression, never authority.</p>
        <p>Agent Compose // canonical palette</p>
      </footer>
      <p class="sr-only" aria-live="polite">${state.copied ? `${escapeHTML(state.copied)} color copied` : ""}</p>
    </main>
  `;
}

async function loadPalette(): Promise<PaletteDocument> {
  const response = await fetch("./palette.json", { cache: "no-store" });
  if (!response.ok) {
    throw new Error(`palette data request failed with ${response.status}`);
  }
  return parsePalette(await response.json());
}

function bindInteractions(data: PaletteDocument): void {
  app.addEventListener("click", (event) => {
    if (!(event.target instanceof Element)) {
      return;
    }
    const button = event.target.closest<HTMLButtonElement>("button[data-action]");
    if (button === null) {
      return;
    }
    if (button.dataset.action === "role" && button.dataset.role !== undefined) {
      state.role = button.dataset.role;
      render(data);
      return;
    }
    if (button.dataset.action === "theme" && button.dataset.theme !== undefined) {
      state.theme = button.dataset.theme === "night" ? "night" : "day";
      render(data);
      return;
    }
    if (
      button.dataset.action === "copy"
      && button.dataset.copy !== undefined
      && button.dataset.key !== undefined
    ) {
      const color = button.dataset.copy;
      const key = button.dataset.key;
      void navigator.clipboard.writeText(color).then(() => {
        state.copied = key;
        render(data);
        window.setTimeout(() => {
          if (state.copied === key) {
            state.copied = "";
            render(data);
          }
        }, 1400);
      });
    }
  });

  app.addEventListener("change", (event) => {
    if (event.target instanceof HTMLSelectElement) {
      state.sort = event.target.value === "alphabetical" ? "alphabetical" : "spectrum";
      render(data);
    }
  });
}

async function start(): Promise<void> {
  const data = await loadPalette();
  ensurePresentationMatches(data);
  bindInteractions(data);
  render(data);
}

void start().catch((error: unknown) => {
  const message = error instanceof Error ? error.message : "unknown palette error";
  app.innerHTML = `<main class="load-error"><h1>Palette unavailable</h1><p>${escapeHTML(message)}</p></main>`;
});
