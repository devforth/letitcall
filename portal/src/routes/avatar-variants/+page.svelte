<script lang="ts">
	// Design sandbox: thirty round fallback-avatar treatments — what a user sees when
	// they have no uploaded picture. Every specimen renders through the real branding
	// tokens, so what you see here is what ships. Pick a number and it gets wired into
	// ui/Avatar.svelte.
	//
	// Four families, in order of how much they commit to:
	//   A  brand-derived  — only --color-primary, safe for every white-label tenant
	//   B  per-person     — colour comes from the email hash, so people differ
	//   C  typographic    — the letterforms do the work
	//   D  generative     — a pattern carries identity
	import Icon from '@iconify/svelte';
	import userIcon from '@iconify-icons/tabler/user';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';

	type Person = { name: string; email: string };

	// Deliberately awkward set: two-word names, a mononym, an email-only account, a
	// three-word name, and a short CJK-latinised name. Letter pairs are spread so the
	// hashed palettes can't look good by luck.
	const people: Person[] = [
		{ name: 'Marcus Chen', email: 'marcus@acme.io' },
		{ name: 'Nina Petrova', email: 'nina.petrova@lumen-studio.com' },
		{ name: 'Sofia Almeida Reyes', email: 'sofia@reyes.consulting' },
		{ name: 'Ada', email: 'ada@solo.dev' },
		{ name: '', email: 'tomas.okafor@northwind.co' },
		{ name: 'Wu Lei', email: 'wu@bytepark.cn' }
	];

	// Same rule as ui/Avatar.svelte, on purpose — including its rough edge on
	// email-only accounts (see the closing note).
	const initialsOf = (p: Person) => {
		const parts = p.name?.trim().split(/\s+/).filter(Boolean) ?? [];
		if (parts.length >= 2) return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
		if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
		return p.email.slice(0, 2).toUpperCase();
	};

	// FNV-1a: cheap, stable across reloads and servers, well spread on short strings.
	const hashOf = (s: string) => {
		let h = 2166136261;
		for (let i = 0; i < s.length; i++) {
			h ^= s.charCodeAt(i);
			h = Math.imul(h, 16777619);
		}
		return h >>> 0;
	};

	// Eight swatches picked to sit beside a green brand without arguing with it:
	// three greens away from the brand, then cool hues, with warm ones kept low-chroma.
	const curated = [174, 200, 224, 262, 292, 22, 44, 128];

	const hslTriplet = (h: number, s: number, l: number) => {
		const a = s * Math.min(l, 1 - l);
		const f = (n: number) => {
			const k = (n + h / 30) % 12;
			const c = l - a * Math.max(-1, Math.min(k - 3, Math.min(9 - k, 1)));
			return Math.round(255 * c);
		};
		return `${f(0)} ${f(8)} ${f(4)}`;
	};

	let size = $state(40);
	let hue = $state(144);
	let hueOn = $state(false);

	// Untouched, the page inherits the tenant's real primary. The slider exists because
	// the brand colour is configurable: a treatment that only works in green isn't done.
	const brandOverride = $derived(hueOn ? `--color-primary: ${hslTriplet(hue, 1, 0.394)};` : '');

	const scaleSteps = [24, 32, 44, 64];

	type Spec = {
		outer: string;
		text?: string;
		textStyle?: string;
		svg?: string;
		inner?: string;
		ghost?: string;
		ghostStyle?: string;
		stack?: boolean;
		icon?: boolean;
	};

	// --- generated layers -------------------------------------------------------
	// currentColor everywhere, so one `color` on the wrapper re-tints the whole layer
	// and the brand slider keeps working.

	// The viewBox is padded so the 5×5 grid sits inside the disc rather than running to
	// the corners, where the round mask would eat whole cells and collapse two different
	// hashes into the same picture.
	const identiconSvg = (seed: number) => {
		let cells = '';
		for (let x = 0; x < 3; x++) {
			for (let y = 0; y < 5; y++) {
				if (!((seed >>> ((x * 5 + y) % 31)) & 1)) continue;
				const o = (seed >>> (x + y + 1)) & 1 ? 0.9 : 0.42;
				cells += `<rect x="${x}" y="${y}" width="1" height="1" rx="0.14" fill="currentColor" opacity="${o}"/>`;
				if (x < 2)
					cells += `<rect x="${4 - x}" y="${y}" width="1" height="1" rx="0.14" fill="currentColor" opacity="${o}"/>`;
			}
		}
		return `<svg viewBox="-0.7 -0.7 6.4 6.4" width="100%" height="100%" aria-hidden="true">${cells}</svg>`;
	};

	// Dots stay in the middle 63% of each axis for the same reason — confetti that only
	// ever lands in the corners is confetti the mask throws away.
	const dotsSvg = (seed: number) => {
		let dots = '';
		for (let i = 0; i < 7; i++) {
			const a = (seed >>> (i * 3)) & 15;
			const b = (seed >>> (i * 3 + 4)) & 15;
			const r = 0.5 + (((seed >>> (i * 2)) & 3) * 0.35);
			dots += `<circle cx="${3 + a * 0.63}" cy="${3 + b * 0.63}" r="${r}" fill="currentColor" opacity="${i % 2 ? 0.5 : 0.28}"/>`;
		}
		return `<svg viewBox="0 0 15 15" width="100%" height="100%" aria-hidden="true">${dots}</svg>`;
	};

	const grain =
		"url(\"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.85' numOctaves='3'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)' opacity='0.5'/%3E%3C/svg%3E\")";

	// --- the thirty -------------------------------------------------------------

	const specFor = (n: number, p: Person, px: number): Spec => {
		const t = initialsOf(p);
		const h = hashOf(p.email.toLowerCase());
		const hu = h % 360;
		const cu = curated[h % curated.length];
		const base = Math.round(px * 0.36);
		const fs = (m = 1) => `font-size: ${Math.round(base * m)}px;`;
		const P = (a?: number) => (a === undefined ? 'rgb(var(--color-primary))' : `rgb(var(--color-primary) / ${a})`);

		switch (n) {
			// ---- A. brand-derived -------------------------------------------------
			case 1:
				return { outer: `background: ${P(0.14)};`, text: t, textStyle: `color: ${P()}; ${fs()}` };
			case 2:
				return {
					outer: `background: ${P()};`,
					text: t,
					textStyle: `color: rgb(var(--color-contrast-text)); ${fs()}`
				};
			case 3:
				return {
					outer: `background: linear-gradient(135deg, ${P(0.35)}, ${P(0.2)} 50%, ${P(0.08)});`,
					text: t,
					textStyle: `color: ${P()}; ${fs()}`
				};
			case 4:
				return {
					outer: `background: linear-gradient(160deg, ${P()}, ${P(0.55)});`,
					text: t,
					textStyle: `color: rgb(var(--color-contrast-text)); ${fs()}`
				};
			case 5:
				return {
					outer: `background: radial-gradient(circle at 32% 24%, ${P(0.38)}, ${P(0.1)} 70%);`,
					text: t,
					textStyle: `color: ${P()}; ${fs()}`
				};
			case 6:
				return {
					outer: `background: transparent; box-shadow: inset 0 0 0 ${Math.max(1, Math.round(px * 0.04))}px ${P(0.45)};`,
					text: t,
					textStyle: `color: ${P()}; ${fs()}`
				};
			case 7:
				return {
					outer: `background: ${P(0.14)}; box-shadow: 0 0 0 2px rgb(var(--color-foreground)), 0 0 0 3px ${P(0.4)};`,
					text: t,
					textStyle: `color: ${P()}; ${fs()}`
				};
			case 8:
				return {
					outer: `background: linear-gradient(180deg, ${P(0.24)}, ${P(0.12)}); box-shadow: inset 0 1px 0 rgb(255 255 255 / 0.5), inset 0 -2px 3px ${P(0.14)};`,
					text: t,
					textStyle: `color: ${P()}; ${fs()}`
				};
			case 9: {
				// Fill level rises with the hash — a quiet per-person marker inside one
				// brand hue. Capped at 25% so the solid band stays clear of the baseline;
				// any higher and it slices through the letters.
				const fill = 14 + (h % 12);
				return {
					outer: `background: linear-gradient(0deg, ${P(0.9)} ${fill}%, ${P(0.12)} ${fill}%);`,
					text: t,
					textStyle: `color: ${P()}; ${fs()}`
				};
			}
			case 10: {
				const arc = 25 + (h % 60);
				return {
					outer: `background: conic-gradient(from -90deg, ${P()} ${arc}%, ${P(0.16)} ${arc}%);`,
					inner: `inset: ${Math.max(2, Math.round(px * 0.09))}px; background: rgb(var(--color-foreground));`,
					text: t,
					textStyle: `color: ${P()}; ${fs(0.92)}`
				};
			}

			// ---- B. per-person colour --------------------------------------------
			case 11:
				return {
					outer: `background: hsl(${hu} 62% 91%);`,
					text: t,
					textStyle: `color: hsl(${hu} 66% 31%); ${fs()}`
				};
			case 12:
				return { outer: `background: hsl(${hu} 54% 45%);`, text: t, textStyle: `color: #fff; ${fs()}` };
			case 13:
				return {
					outer: `background: linear-gradient(135deg, hsl(${hu} 62% 52%), hsl(${(hu + 34) % 360} 66% 42%));`,
					text: t,
					textStyle: `color: #fff; ${fs()}`
				};
			case 14:
				return {
					outer: `background: hsl(${cu} 58% 90%);`,
					text: t,
					textStyle: `color: hsl(${cu} 62% 29%); ${fs()}`
				};
			case 15:
				return {
					outer: `background: hsl(${cu} 58% 90%); box-shadow: 0 0 0 ${Math.max(1, Math.round(px * 0.05))}px ${P(0.55)};`,
					text: t,
					textStyle: `color: hsl(${cu} 62% 29%); ${fs()}`
				};
			case 16:
				return {
					outer: `background: hsl(${hu} 40% 94%); background-image: radial-gradient(circle at 22% 18%, hsl(${hu} 78% 72% / 0.95), transparent 58%), radial-gradient(circle at 78% 82%, hsl(${(hu + 46) % 360} 78% 68% / 0.9), transparent 60%);`,
					text: t,
					textStyle: `color: hsl(${hu} 62% 24%); ${fs()}`
				};

			// ---- C. typographic --------------------------------------------------
			case 17:
				return { outer: `background: ${P(0.14)};`, text: t[0], textStyle: `color: ${P()}; ${fs(1.45)}` };
			case 18:
				return {
					outer: `background: ${P(0.12)};`,
					text: t.toLowerCase(),
					textStyle: `color: ${P()}; ${fs(0.92)} font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; letter-spacing: 0.04em; font-weight: 600;`
				};
			case 19:
				return {
					outer: `background: ${P(0.1)};`,
					text: t[0],
					textStyle: `color: ${P()}; ${fs(1.6)} font-family: Georgia, 'Iowan Old Style', 'Times New Roman', serif; font-weight: 400;`
				};
			case 20:
				return {
					outer: `background: ${P(0.1)};`,
					text: t,
					textStyle: `color: transparent; -webkit-text-stroke: ${Math.max(0.8, px * 0.032)}px ${P(0.95)}; ${fs(1.05)} font-weight: 800;`
				};
			case 21:
				return {
					outer: `background: ${P(0.12)};`,
					ghost: t[0],
					ghostStyle: `color: ${P(0.28)}; font-size: ${Math.round(px * 0.92)}px; font-weight: 800; right: -6%; bottom: -22%;`,
					text: t,
					textStyle: `color: ${P()}; ${fs(0.85)}`
				};
			case 22:
				return {
					outer: `background: ${P(0.14)};`,
					text: t,
					stack: true,
					textStyle: `color: ${P()}; ${fs(0.62)} line-height: 0.94; letter-spacing: 0.02em;`
				};

			// ---- D. generative ---------------------------------------------------
			case 23:
				return { outer: `background: ${P(0.1)}; color: ${P()};`, svg: identiconSvg(h) };
			case 24:
				return {
					outer: `background: ${P(0.1)}; color: ${P(0.3)};`,
					svg: identiconSvg(h),
					text: t,
					textStyle: `color: ${P()}; ${fs()}`
				};
			case 25: {
				const step = Math.max(3, Math.round(px * (0.1 + (h % 3) * 0.03)));
				return {
					outer: `background: repeating-radial-gradient(circle at 50% 50%, ${P(0.26)} 0 ${step * 0.5}px, ${P(0.08)} ${step * 0.5}px ${step}px);`,
					text: t,
					textStyle: `color: ${P()}; ${fs()}`
				};
			}
			case 26:
				return {
					outer: `background: ${P(0.1)}; color: ${P()};`,
					svg: dotsSvg(h),
					text: t,
					textStyle: `color: ${P()}; ${fs()}`
				};
			case 27:
				return {
					outer: `background: ${P(0.1)}; background-image: repeating-linear-gradient(45deg, ${P(0.16)} 0 ${Math.max(2, px * 0.07)}px, transparent ${Math.max(2, px * 0.07)}px ${Math.max(4, px * 0.14)}px);`,
					text: t,
					textStyle: `color: ${P()}; ${fs()}`
				};
			case 28:
				return {
					outer: `background-color: ${P(0.07)}; background-image: radial-gradient(${P(0.4)} 1px, transparent 1px); background-size: ${Math.max(3, Math.round(px * 0.1))}px ${Math.max(3, Math.round(px * 0.1))}px;`,
					text: t,
					textStyle: `color: ${P()}; ${fs()}`
				};
			case 29:
				return {
					outer: `background-color: ${P(0.16)}; background-image: ${grain}; background-blend-mode: overlay;`,
					text: t,
					textStyle: `color: ${P()}; ${fs()}`
				};
			case 30:
				return { outer: `background: ${P(0.14)}; color: ${P()};`, icon: true };
			default:
				return { outer: `background: ${P(0.14)};`, text: t, textStyle: `color: ${P()}; ${fs()}` };
		}
	};

	type Variant = { n: number; title: string; note: string; tags: string[] };

	const families: { key: string; title: string; blurb: string; variants: Variant[] }[] = [
		{
			key: 'brand',
			title: 'A · Brand-derived',
			blurb:
				'Built from --color-primary alone, so every tenant gets their own colour and nothing has to be maintained per-hue. The trade: two people in a list look alike apart from their letters.',
			variants: [
				{ n: 1, title: 'Soft tint', note: 'Flat 14% brand wash, brand letters. The quiet default.', tags: ['1 property'] },
				{ n: 2, title: 'Solid brand', note: 'Full-strength fill with contrast text. Loudest option — carries weight in a header.', tags: ['high contrast'] },
				{ n: 3, title: 'Diagonal tint', note: 'What ships today: three brand stops on a 135° ramp.', tags: ['current'] },
				{ n: 4, title: 'Deep gradient', note: 'Solid brand into 55%, so the fill has direction without going pale.', tags: ['high contrast'] },
				{ n: 5, title: 'Radial spotlight', note: 'Off-centre light source — reads as a sphere rather than a disc.', tags: ['dimensional'] },
				{ n: 6, title: 'Ghost outline', note: 'No fill at all. Disappears into dense tables, which is sometimes the point.', tags: ['lightest'] },
				{ n: 7, title: 'Halo ring', note: 'Tint plus a detached outer ring. Separates avatars that overlap in a stack.', tags: ['needs padding'] },
				{ n: 8, title: 'Bevel', note: 'Top highlight, bottom inner shadow. Physical, slightly retro.', tags: ['dimensional'] },
				{ n: 9, title: 'Fill level', note: 'Hash sets how high the solid brand rises. Same hue, still individual.', tags: ['per-person'] },
				{ n: 10, title: 'Arc ring', note: 'Conic arc around a hollow core. Doubles as a slot for a real metric later.', tags: ['extensible'] }
			]
		},
		{
			key: 'person',
			title: 'B · Per-person colour',
			blurb:
				'Colour derives from the email hash, so a face in a list is recognisable before you read it. Costs you the guarantee that avatars stay on-brand — worth it in the users table, risky next to the logo.',
			variants: [
				{ n: 11, title: 'Hashed pastel', note: 'Any hue, held at one lightness so contrast never drifts.', tags: ['360 hues'] },
				{ n: 12, title: 'Hashed solid', note: 'Saturated fill, white text. Punchy; a few hues will clash with green.', tags: ['360 hues'] },
				{ n: 13, title: 'Hashed duo', note: 'Hue into hue+34 — the ramp reads as intentional, not random.', tags: ['360 hues'] },
				{ n: 14, title: 'Curated eight', note: 'Fixed swatch list chosen against the brand. Predictable, never ugly.', tags: ['8 swatches', 'safest'] },
				{ n: 15, title: 'Curated + brand ring', note: 'Person in the fill, brand in the ring. Has it both ways.', tags: ['8 swatches'] },
				{ n: 16, title: 'Mesh blobs', note: 'Two soft radial blobs. Modern, and the heaviest thing here to render.', tags: ['expensive'] }
			]
		},
		{
			key: 'type',
			title: 'C · Typographic',
			blurb:
				'One tint, and the letterforms carry the character. These are the cheapest to ship and the most sensitive to size — check them at 24px in the strip under each specimen.',
			variants: [
				{ n: 17, title: 'Single letter', note: 'One initial, oversized. Confident, and it drops the surname signal.', tags: ['loses a letter'] },
				{ n: 18, title: 'Mono lowercase', note: 'Lowercase monospace. Technical, quietly friendly.', tags: ['system mono'] },
				{ n: 19, title: 'Serif initial', note: 'Light serif against a sans UI. Editorial; the sharpest departure on the page.', tags: ['system serif'] },
				{ n: 20, title: 'Outlined', note: 'Stroked letters, hollow centres. Falls apart below ~32px.', tags: ['-webkit only'] },
				{ n: 21, title: 'Ghost letter', note: 'A huge initial, cropped by the disc, sits behind the normal pair.', tags: ['two layers'] },
				{ n: 22, title: 'Stacked', note: 'Initials stacked instead of set side by side. Compact, monogram-ish.', tags: ['tight'] }
			]
		},
		{
			key: 'gen',
			title: 'D · Generative',
			blurb:
				'A pattern seeded by the email. Strongest identity signal available without a photo, and the only family where the shape itself is the name — so it needs the most testing at small sizes.',
			variants: [
				{ n: 23, title: 'Identicon', note: 'Mirrored 5×5 grid, no letters at all. Distinct even at 24px.', tags: ['no letters', 'svg'] },
				{ n: 24, title: 'Identicon + initials', note: 'Grid dropped to 30% behind the letters. Both signals, one disc.', tags: ['svg'] },
				{ n: 25, title: 'Concentric rings', note: 'Hash sets ring spacing. Cheap — one repeating gradient.', tags: ['1 property'] },
				{ n: 26, title: 'Confetti', note: 'Seven seeded dots behind the letters. Playful; watch it fight the type.', tags: ['svg'] },
				{ n: 27, title: 'Diagonal stripes', note: 'Fixed 45° hatch. Texture without any per-person meaning.', tags: ['decorative'] },
				{ n: 28, title: 'Halftone dots', note: 'Dot grid, brand-tinted. Print-shop texture at UI scale.', tags: ['decorative'] },
				{ n: 29, title: 'Grain', note: 'Turbulence overlay on a brand wash. Warms the flat tint up.', tags: ['inline filter'] },
				{ n: 30, title: 'Icon fallback', note: 'A person glyph instead of letters — the honest answer when you have neither.', tags: ['no letters'] }
			]
		}
	];
</script>

<svelte:head><title>Avatar variants</title></svelte:head>

<main class="page" style={brandOverride}>
	<header class="page-head">
		<div class="head-text">
			<h1 class="title">Initials avatars — 30 treatments</h1>
			<p class="sub">
				Every fallback avatar a user can get when they haven't uploaded a picture. Specimens sit on
				<code>--color-foreground</code> and read the live branding tokens, so this is the real thing, not a mock.
				Tell me a number and it becomes <code>Avatar.svelte</code>.
			</p>
		</div>
		<ThemeToggle />
	</header>

	<div class="toolbar">
		<label class="control">
			<span class="control-label">Size <b>{size}px</b></span>
			<input type="range" min="24" max="72" step="2" bind:value={size} />
		</label>
		<label class="control">
			<span class="control-label">
				Brand hue {hueOn ? hue + '°' : '— tenant brand'}
			</span>
			<input
				type="range"
				min="0"
				max="359"
				bind:value={hue}
				oninput={() => (hueOn = true)}
			/>
		</label>
		<button type="button" class="reset" disabled={!hueOn} onclick={() => { hueOn = false; hue = 144; }}>
			Back to brand
		</button>
	</div>

	<p class="legend">
		{#each people as p (p.email)}
			<span class="legend-item">{p.name || p.email}</span>
		{/each}
	</p>

	{#each families as f (f.key)}
		<section class="family">
			<div class="family-head">
				<h2 class="family-title">{f.title}</h2>
				<p class="family-blurb">{f.blurb}</p>
			</div>

			<div class="grid">
				{#each f.variants as v (v.n)}
					<article class="card">
						<div class="card-head">
							<span class="badge">{v.n}</span>
							<div>
								<h3 class="card-title">{v.title}</h3>
								<p class="card-note">{v.note}</p>
							</div>
						</div>

						<div class="row" style="min-height: {size + 8}px;">
							{#each people as p (p.email)}
								{@const s = specFor(v.n, p, size)}
								<span
									class="av"
									title={p.name || p.email}
									aria-label={p.name || p.email}
									style="width: {size}px; height: {size}px; {s.outer}"
								>
									{#if s.svg || s.ghost}
										<span class="av-mask">
											{#if s.svg}{@html s.svg}{/if}
											{#if s.ghost}<span class="av-ghost" style={s.ghostStyle}>{s.ghost}</span>{/if}
										</span>
									{/if}
									{#if s.inner}<span class="av-inner" style={s.inner}></span>{/if}
									{#if s.icon}
										<Icon icon={userIcon} width={Math.round(size * 0.55)} height={Math.round(size * 0.55)} />
									{:else if s.text}
										<span class="av-text" class:av-stack={s.stack} style={s.textStyle}>
											{#if s.stack}{#each s.text.split('') as ch, i (i)}<span>{ch}</span>{/each}{:else}{s.text}{/if}
										</span>
									{/if}
								</span>
							{/each}
						</div>

						<div class="scale">
							{#each scaleSteps as px (px)}
								{@const s = specFor(v.n, people[0], px)}
								<span class="scale-cell">
									<span class="av" style="width: {px}px; height: {px}px; {s.outer}">
										{#if s.svg || s.ghost}
											<span class="av-mask">
												{#if s.svg}{@html s.svg}{/if}
												{#if s.ghost}<span class="av-ghost" style={s.ghostStyle}>{s.ghost}</span>{/if}
											</span>
										{/if}
										{#if s.inner}<span class="av-inner" style={s.inner}></span>{/if}
										{#if s.icon}
											<Icon icon={userIcon} width={Math.round(px * 0.55)} height={Math.round(px * 0.55)} />
										{:else if s.text}
											<span class="av-text" class:av-stack={s.stack} style={s.textStyle}>
												{#if s.stack}{#each s.text.split('') as ch, i (i)}<span>{ch}</span>{/each}{:else}{s.text}{/if}
											</span>
										{/if}
									</span>
									<span class="scale-px">{px}</span>
								</span>
							{/each}
							<span class="tags">
								{#each v.tags as tag (tag)}<span class="tag">{tag}</span>{/each}
							</span>
						</div>
					</article>
				{/each}
			</div>
		</section>
	{/each}

	<footer class="closing">
		<h2 class="family-title">Notes before you pick</h2>
		<ul>
			<li>
				<b>Real sizes in the app are 22–44px.</b> Avatar is called at 22, 24, 34, 36, 40 and 44px across
				<code>HostBadges</code>, <code>HostSelector</code>, <code>UserTable</code>, <code>AppShell</code> and
				<code>AuditLogTable</code>. The 24px column in each strip is the one that decides — variants 20 and 23
				are the two that change character most between 24 and 64.
			</li>
			<li>
				<b>Family B breaks white-label.</b> Hashed hues ignore <code>--color-primary</code> entirely. If tenants
				expect their brand everywhere, 14 or 15 are the compromises; 15 keeps a brand ring around a per-person fill.
			</li>
			<li>
				<b>Dark mode ground is light.</b> <code>--color-foreground</code> is <code>100 100 100</code> in dark, so
				low-alpha brand tints (1, 25, 28, 29) lose more contrast there than they do on white. Flip the theme
				before committing to anything under ~12% alpha.
			</li>
			<li>
				<b>Email-only accounts get poor initials today.</b> <code>tomas.okafor@northwind.co</code> renders as
				<code>TO</code> here only because the local part starts with two letters — a real address like
				<code>t.okafor@…</code> yields <code>T.</code>, punctuation and all. Worth fixing in
				<code>Avatar.svelte</code> alongside whichever look you choose: split the local part on
				<code>.</code>/<code>_</code>/<code>-</code> and take the first letter of each piece.
			</li>
		</ul>
	</footer>
</main>

<style>
	.page {
		max-width: 78rem;
		margin: 0 auto;
		padding: 2rem 1.5rem 5rem;
		color: rgb(var(--color-text));
	}

	.page-head {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1.5rem;
		margin-bottom: 1.5rem;
	}

	.head-text {
		max-width: 46rem;
	}

	.title {
		font-size: 1.5rem;
		font-weight: 600;
		letter-spacing: -0.015em;
		text-wrap: balance;
		color: rgb(var(--color-text));
	}

	.sub {
		margin-top: 0.5rem;
		font-size: 0.875rem;
		line-height: 1.55;
		opacity: 0.72;
	}

	code {
		font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
		font-size: 0.9em;
		white-space: nowrap;
		padding: 0.05em 0.3em;
		border-radius: 0.25rem;
		background: rgb(var(--color-primary) / 0.1);
		color: rgb(var(--color-primary));
	}

	/* ---------- toolbar ---------- */

	.toolbar {
		position: sticky;
		top: 0;
		z-index: 5;
		display: flex;
		flex-wrap: wrap;
		align-items: flex-end;
		gap: 1.5rem;
		margin-bottom: 1rem;
		padding: 0.875rem 1.125rem;
		border: 1px solid rgb(var(--color-border));
		border-radius: 0.875rem;
		background: rgb(var(--color-foreground));
		box-shadow: var(--shadow-small);
	}

	.control {
		display: grid;
		gap: 0.375rem;
		min-width: 13rem;
	}

	.control-label {
		font-size: 0.6875rem;
		font-weight: 700;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		opacity: 0.55;
	}

	.control-label b {
		font-variant-numeric: tabular-nums;
		opacity: 0.9;
	}

	input[type='range'] {
		width: 100%;
		accent-color: rgb(var(--color-primary));
	}

	input[type='range']:focus-visible {
		outline: 2px solid rgb(var(--color-primary));
		outline-offset: 3px;
	}

	.reset {
		padding: 0.4rem 0.75rem;
		border-radius: 0.5rem;
		border: 1px solid rgb(var(--color-border));
		background: transparent;
		color: inherit;
		font-size: 0.75rem;
		font-weight: 600;
		cursor: pointer;
	}

	.reset:hover:not(:disabled) {
		background: rgb(var(--color-primary) / 0.1);
		border-color: rgb(var(--color-primary) / 0.4);
		color: rgb(var(--color-primary));
	}

	.reset:disabled {
		opacity: 0.4;
		cursor: default;
	}

	.reset:focus-visible {
		outline: 2px solid rgb(var(--color-primary));
		outline-offset: 2px;
	}

	.legend {
		display: flex;
		flex-wrap: wrap;
		gap: 0.375rem 1rem;
		margin-bottom: 2.5rem;
		font-size: 0.6875rem;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		opacity: 0.5;
	}

	.legend-item + .legend-item::before {
		content: '·';
		margin-right: 1rem;
	}

	/* ---------- families ---------- */

	.family {
		margin-bottom: 3rem;
	}

	.family-head {
		max-width: 52rem;
		margin-bottom: 1.25rem;
		padding-bottom: 0.875rem;
		border-bottom: 1px solid rgb(var(--color-border));
	}

	.family-title {
		font-size: 1rem;
		font-weight: 700;
		letter-spacing: 0.01em;
		color: rgb(var(--color-text));
	}

	.family-blurb {
		margin-top: 0.4rem;
		font-size: 0.8125rem;
		line-height: 1.55;
		opacity: 0.68;
	}

	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(21rem, 1fr));
		gap: 1rem;
	}

	.card {
		display: flex;
		flex-direction: column;
		gap: 1rem;
		border: 1px solid rgb(var(--color-border));
		border-radius: 1rem;
		background: rgb(var(--color-foreground));
		box-shadow: var(--shadow-small);
		padding: 1.125rem 1.25rem 1rem;
	}

	.card-head {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
	}

	.badge {
		display: grid;
		place-items: center;
		flex-shrink: 0;
		width: 1.75rem;
		height: 1.75rem;
		border-radius: 0.5rem;
		background: rgb(var(--color-primary) / 0.14);
		color: rgb(var(--color-primary));
		font-size: 0.8125rem;
		font-weight: 700;
		font-variant-numeric: tabular-nums;
	}

	.card-title {
		font-size: 0.9375rem;
		font-weight: 600;
		color: rgb(var(--color-text));
	}

	.card-note {
		margin-top: 0.25rem;
		font-size: 0.75rem;
		line-height: 1.5;
		opacity: 0.65;
	}

	/* ---------- specimens ---------- */

	.row {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.625rem;
	}

	.av {
		position: relative;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		overflow: hidden;
		border-radius: 9999px;
		font-weight: 700;
		line-height: 1;
	}

	/* Generated layers get their own round mask. clip-path does the work — an
	   overflow: hidden on .av alone leaves absolutely-positioned children spilling
	   into the corners of the box, outside the disc. */
	.av-mask {
		position: absolute;
		inset: 0;
		display: block;
		line-height: 0;
		border-radius: 9999px;
		overflow: hidden;
		clip-path: circle(50%);
	}

	.av-inner {
		position: absolute;
		border-radius: 9999px;
	}

	.av-ghost {
		position: absolute;
		line-height: 0.8;
		pointer-events: none;
		user-select: none;
	}

	.av-text {
		position: relative;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		white-space: nowrap;
	}

	.av-stack {
		flex-direction: column;
	}

	.scale {
		display: flex;
		align-items: flex-end;
		gap: 0.75rem;
		margin-top: auto;
		padding-top: 0.875rem;
		border-top: 1px solid rgb(var(--color-border));
	}

	.scale-cell {
		display: grid;
		justify-items: center;
		gap: 0.25rem;
	}

	.scale-px {
		font-size: 0.625rem;
		font-variant-numeric: tabular-nums;
		opacity: 0.4;
	}

	.tags {
		display: flex;
		flex-wrap: wrap;
		gap: 0.25rem;
		margin-left: auto;
		justify-content: flex-end;
	}

	.tag {
		padding: 0.125rem 0.4rem;
		border-radius: 999px;
		border: 1px solid rgb(var(--color-border));
		font-size: 0.625rem;
		font-weight: 600;
		letter-spacing: 0.02em;
		opacity: 0.6;
		white-space: nowrap;
	}

	/* ---------- closing ---------- */

	.closing {
		max-width: 52rem;
		padding-top: 1.5rem;
		border-top: 1px solid rgb(var(--color-border));
	}

	.closing ul {
		display: grid;
		gap: 0.75rem;
		margin-top: 0.875rem;
		padding-left: 1.1rem;
	}

	.closing li {
		font-size: 0.8125rem;
		line-height: 1.6;
		opacity: 0.75;
	}

	.closing b {
		font-weight: 600;
		opacity: 1;
	}

	@media (max-width: 640px) {
		.page-head {
			flex-direction: column;
		}

		.grid {
			grid-template-columns: 1fr;
		}
	}
</style>
