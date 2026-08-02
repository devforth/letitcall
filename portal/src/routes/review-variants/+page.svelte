<script lang="ts">
	// Design sandbox: ten looks for the booking flow's confirmation (review) step,
	// each fed the same booking. Pick one and it gets wired into
	// src/routes/book/[slug]/+page.svelte.
	import Icon from '@iconify/svelte';
	import calendarIcon from '@iconify-icons/tabler/calendar';
	import clockIcon from '@iconify-icons/tabler/clock';
	import worldIcon from '@iconify-icons/tabler/world';
	import userIcon from '@iconify-icons/tabler/user';
	import mailIcon from '@iconify-icons/tabler/mail';
	import usersIcon from '@iconify-icons/tabler/users';
	import notesIcon from '@iconify-icons/tabler/align-left';
	import pencilIcon from '@iconify-icons/tabler/pencil';
	import videoIcon from '@iconify-icons/tabler/video';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';
	import { page } from '$app/state';

	const booking = {
		event: 'Product demo',
		duration: 30,
		host: { name: 'Danny Kim', email: 'danny@letitcall.app' },
		dateLong: 'Friday, August 15, 2026',
		dateShort: 'Fri, Aug 15',
		weekday: 'Friday',
		day: '15',
		month: 'Aug',
		range: '10:00 AM – 10:30 AM',
		timezone: 'Europe/Kyiv',
		name: 'Alona Radchenko',
		email: 'alona@devforth.io',
		guests: [
			'nina@acme.io',
			'marcus.long.name@company-with-a-long-domain.com',
			'dev@letitcall.app'
		],
		notes: 'Want to walk through the onboarding flow and the pricing page copy.'
	};

	const guestCount = booking.guests.length;

	const variants = [
		{ n: 1, title: 'Icon · label · value rows', note: 'Current build: brand glyph, fixed label column, value beside it. Guests listed one per line.', wide: false },
		{ n: 2, title: 'Definition list', note: 'No icons. Micro-caps labels left, values right-aligned, hairline between rows — reads like a receipt.', wide: false },
		{ n: 3, title: 'Ticket', note: 'Time and date as a filled header, perforated tear line, details in the stub below.', wide: false },
		{ n: 4, title: 'Fact cards', note: 'Every fact is its own tinted card in a 2-up grid; guests and notes span the full width.', wide: false },
		{ n: 5, title: 'Timeline rail', note: 'Icon bubbles on a vertical rail — same visual language as the stepper above it.', wide: false },
		{ n: 6, title: 'Sentence summary', note: 'One human sentence carries when/who; guests become avatar chips, notes a quote block.', wide: false },
		{ n: 7, title: 'Editable rows', note: 'Striped rows with a per-row Edit link that jumps back to the step that owns the field.', wide: false },
		{ n: 8, title: 'When / Who split', note: 'Two panels side by side: the slot on the left, the people on the right, notes underneath.', wide: true },
		{ n: 9, title: 'Dense meta lines', note: 'Shortest possible: middot-separated meta lines, guests as inline chips. Fits without scrolling.', wide: false },
		{ n: 10, title: 'Hero date', note: 'Oversized day number anchors the block; everything else is muted meta. Full-width confirm.', wide: false }
	];

	// ?only=3 isolates one variant, which makes it easy to grab a clean screenshot.
	const only = $derived(Number(page.url.searchParams.get('only')) || null);
	const shown = $derived(only ? variants.filter((v) => v.n === only) : variants);
</script>

<svelte:head><title>Booking review variants</title></svelte:head>

<main class="page">
	<header class="page-head">
		<div>
			<h1 class="text-2xl font-semibold">Booking review — 10 variants</h1>
			<p class="mt-2 text-sm opacity-70">
				Same booking in every card, step 3 of the booking flow. Tell me a number and I'll make it the real one.
			</p>
		</div>
		<ThemeToggle />
	</header>

	<div class="grid gap-6">
		{#each shown as v (v.n)}
			<article class="card">
				<div class="card-head">
					<span class="badge">{v.n}</span>
					<div>
						<h2 class="text-base font-semibold">{v.title}</h2>
						<p class="mt-1 text-xs opacity-65">{v.note}</p>
					</div>
				</div>

				<div class="frame" class:is-wide={v.wide}>
					<!-- ============================== 1 ============================== -->
					{#if v.n === 1}
						<h3 class="text-lg font-semibold">Review your booking</h3>
						<div class="mt-5 grid gap-4 text-sm">
							<div class="r1-row">
								<Icon icon={calendarIcon} width="20" height="20" class="r1-icon" />
								<p class="r1-label">Date</p>
								<p class="r1-value">{booking.dateLong}</p>
							</div>
							<div class="r1-row">
								<Icon icon={clockIcon} width="20" height="20" class="r1-icon" />
								<p class="r1-label">Time</p>
								<p class="r1-value">{booking.range}</p>
							</div>
							<div class="r1-row">
								<Icon icon={worldIcon} width="20" height="20" class="r1-icon" />
								<p class="r1-label">Timezone</p>
								<p class="r1-value">{booking.timezone}</p>
							</div>
							<div class="r1-row">
								<Icon icon={userIcon} width="20" height="20" class="r1-icon" />
								<p class="r1-label">Name</p>
								<p class="r1-value">{booking.name}</p>
							</div>
							<div class="r1-row">
								<Icon icon={mailIcon} width="20" height="20" class="r1-icon" />
								<p class="r1-label">Email</p>
								<p class="r1-value">{booking.email}</p>
							</div>
							<div class="r1-row">
								<Icon icon={usersIcon} width="20" height="20" class="r1-icon" />
								<p class="r1-label">Guests <span class="r1-count">{guestCount}</span></p>
								<ul class="r1-guests">
									{#each booking.guests as guest (guest)}
										<li class="r1-value">{guest}</li>
									{/each}
								</ul>
							</div>
							<div class="r1-row">
								<Icon icon={notesIcon} width="20" height="20" class="r1-icon" />
								<p class="r1-label">Notes</p>
								<p class="r1-value">{booking.notes}</p>
							</div>
						</div>
						<div class="mt-6"><Button>Confirm booking</Button></div>

						<!-- ============================== 2 ============================== -->
					{:else if v.n === 2}
						<h3 class="text-lg font-semibold">Review your booking</h3>
						<dl class="r2">
							<div class="r2-row"><dt>Date</dt><dd>{booking.dateLong}</dd></div>
							<div class="r2-row"><dt>Time</dt><dd>{booking.range}</dd></div>
							<div class="r2-row"><dt>Timezone</dt><dd>{booking.timezone}</dd></div>
							<div class="r2-row"><dt>Name</dt><dd>{booking.name}</dd></div>
							<div class="r2-row"><dt>Email</dt><dd>{booking.email}</dd></div>
							<div class="r2-row">
								<dt>Guests · {guestCount}</dt>
								<dd>
									<ul class="r2-guests">
										{#each booking.guests as guest (guest)}
											<li>{guest}</li>
										{/each}
									</ul>
								</dd>
							</div>
							<div class="r2-row"><dt>Notes</dt><dd>{booking.notes}</dd></div>
						</dl>
						<div class="mt-6"><Button>Confirm booking</Button></div>

						<!-- ============================== 3 ============================== -->
					{:else if v.n === 3}
						<div class="r3">
							<div class="r3-head">
								<p class="r3-event">{booking.event} · {booking.duration} min</p>
								<p class="r3-time">{booking.range}</p>
								<p class="r3-date">{booking.dateLong}</p>
								<p class="r3-tz"><Icon icon={worldIcon} width="15" height="15" />{booking.timezone}</p>
							</div>
							<div class="r3-perf" aria-hidden="true"></div>
							<div class="r3-body">
								<div>
									<p class="r3-k">Booked by</p>
									<p class="r3-v">{booking.name}</p>
									<p class="r3-sub">{booking.email}</p>
								</div>
								<div>
									<p class="r3-k">Host</p>
									<p class="r3-v">{booking.host.name}</p>
									<p class="r3-sub">{booking.host.email}</p>
								</div>
								<div class="r3-full">
									<p class="r3-k">Guests · {guestCount}</p>
									<ul class="r3-guests">
										{#each booking.guests as guest (guest)}
											<li>{guest}</li>
										{/each}
									</ul>
								</div>
								<div class="r3-full">
									<p class="r3-k">Notes</p>
									<p class="r3-v r3-notes">{booking.notes}</p>
								</div>
							</div>
						</div>
						<div class="mt-6"><Button>Confirm booking</Button></div>

						<!-- ============================== 4 ============================== -->
					{:else if v.n === 4}
						<h3 class="text-lg font-semibold">Review your booking</h3>
						<div class="r4">
							<div class="r4-card">
								<p class="r4-k"><Icon icon={calendarIcon} width="16" height="16" />Date</p>
								<p class="r4-v">{booking.dateLong}</p>
							</div>
							<div class="r4-card">
								<p class="r4-k"><Icon icon={clockIcon} width="16" height="16" />Time</p>
								<p class="r4-v">{booking.range}</p>
							</div>
							<div class="r4-card">
								<p class="r4-k"><Icon icon={worldIcon} width="16" height="16" />Timezone</p>
								<p class="r4-v">{booking.timezone}</p>
							</div>
							<div class="r4-card">
								<p class="r4-k"><Icon icon={userIcon} width="16" height="16" />You</p>
								<p class="r4-v">{booking.name}</p>
								<p class="r4-sub">{booking.email}</p>
							</div>
							<div class="r4-card r4-full">
								<p class="r4-k"><Icon icon={usersIcon} width="16" height="16" />Guests · {guestCount}</p>
								<ul class="r4-guests">
									{#each booking.guests as guest (guest)}
										<li>
											<Avatar name={guest} email={guest} size={26} rounded="full" />
											<span class="truncate">{guest}</span>
										</li>
									{/each}
								</ul>
							</div>
							<div class="r4-card r4-full">
								<p class="r4-k"><Icon icon={notesIcon} width="16" height="16" />Notes</p>
								<p class="r4-v r4-notes">{booking.notes}</p>
							</div>
						</div>
						<div class="mt-6"><Button>Confirm booking</Button></div>

						<!-- ============================== 5 ============================== -->
					{:else if v.n === 5}
						<h3 class="text-lg font-semibold">Review your booking</h3>
						<ul class="r5">
							<li class="r5-item">
								<span class="r5-bubble"><Icon icon={calendarIcon} width="16" height="16" /></span>
								<div>
									<p class="r5-k">When</p>
									<p class="r5-v">{booking.dateLong}</p>
									<p class="r5-sub">{booking.range} · {booking.timezone}</p>
								</div>
							</li>
							<li class="r5-item">
								<span class="r5-bubble"><Icon icon={videoIcon} width="16" height="16" /></span>
								<div>
									<p class="r5-k">What</p>
									<p class="r5-v">{booking.event}</p>
									<p class="r5-sub">{booking.duration} min with {booking.host.name}</p>
								</div>
							</li>
							<li class="r5-item">
								<span class="r5-bubble"><Icon icon={userIcon} width="16" height="16" /></span>
								<div>
									<p class="r5-k">You</p>
									<p class="r5-v">{booking.name}</p>
									<p class="r5-sub">{booking.email}</p>
								</div>
							</li>
							<li class="r5-item">
								<span class="r5-bubble"><Icon icon={usersIcon} width="16" height="16" /></span>
								<div>
									<p class="r5-k">Guests · {guestCount}</p>
									<ul class="r5-guests">
										{#each booking.guests as guest (guest)}
											<li class="r5-v">{guest}</li>
										{/each}
									</ul>
								</div>
							</li>
							<li class="r5-item r5-last">
								<span class="r5-bubble"><Icon icon={notesIcon} width="16" height="16" /></span>
								<div>
									<p class="r5-k">Notes</p>
									<p class="r5-v r5-notes">{booking.notes}</p>
								</div>
							</li>
						</ul>
						<div class="mt-6"><Button>Confirm booking</Button></div>

						<!-- ============================== 6 ============================== -->
					{:else if v.n === 6}
						<p class="r6-lede">
							You're booking <strong>{booking.event}</strong> with <strong>{booking.host.name}</strong> on
							<strong>{booking.dateShort}</strong> at <strong>{booking.range}</strong>.
						</p>
						<p class="r6-meta">{booking.duration} min · {booking.timezone} · confirmation goes to {booking.email}</p>
						<div class="r6-block">
							<p class="r6-k">Guests · {guestCount}</p>
							<ul class="r6-chips">
								{#each booking.guests as guest (guest)}
									<li class="r6-chip">
										<Avatar name={guest} email={guest} size={22} rounded="full" />
										<span class="truncate">{guest}</span>
									</li>
								{/each}
							</ul>
						</div>
						<blockquote class="r6-quote">{booking.notes}</blockquote>
						<div class="mt-6"><Button>Confirm booking</Button></div>

						<!-- ============================== 7 ============================== -->
					{:else if v.n === 7}
						<h3 class="text-lg font-semibold">Review your booking</h3>
						<ul class="r7">
							<li class="r7-row">
								<Icon icon={calendarIcon} width="18" height="18" class="r7-icon" />
								<div><p class="r7-k">Date &amp; time</p><p class="r7-v">{booking.dateShort}, {booking.range}</p></div>
								<button type="button" class="r7-edit"><Icon icon={pencilIcon} width="14" height="14" />Edit</button>
							</li>
							<li class="r7-row">
								<Icon icon={worldIcon} width="18" height="18" class="r7-icon" />
								<div><p class="r7-k">Timezone</p><p class="r7-v">{booking.timezone}</p></div>
								<button type="button" class="r7-edit"><Icon icon={pencilIcon} width="14" height="14" />Edit</button>
							</li>
							<li class="r7-row">
								<Icon icon={userIcon} width="18" height="18" class="r7-icon" />
								<div><p class="r7-k">You</p><p class="r7-v">{booking.name} · {booking.email}</p></div>
								<button type="button" class="r7-edit"><Icon icon={pencilIcon} width="14" height="14" />Edit</button>
							</li>
							<li class="r7-row">
								<Icon icon={usersIcon} width="18" height="18" class="r7-icon" />
								<div>
									<p class="r7-k">Guests · {guestCount}</p>
									<ul class="r7-guests">
										{#each booking.guests as guest (guest)}
											<li class="r7-v">{guest}</li>
										{/each}
									</ul>
								</div>
								<button type="button" class="r7-edit"><Icon icon={pencilIcon} width="14" height="14" />Edit</button>
							</li>
							<li class="r7-row">
								<Icon icon={notesIcon} width="18" height="18" class="r7-icon" />
								<div><p class="r7-k">Notes</p><p class="r7-v">{booking.notes}</p></div>
								<button type="button" class="r7-edit"><Icon icon={pencilIcon} width="14" height="14" />Edit</button>
							</li>
						</ul>
						<div class="mt-6"><Button>Confirm booking</Button></div>

						<!-- ============================== 8 ============================== -->
					{:else if v.n === 8}
						<h3 class="text-lg font-semibold">Review your booking</h3>
						<div class="r8">
							<section class="r8-panel">
								<p class="r8-head"><Icon icon={calendarIcon} width="16" height="16" />When</p>
								<p class="r8-day">{booking.weekday}</p>
								<p class="r8-date">{booking.month} {booking.day}, 2026</p>
								<p class="r8-time">{booking.range}</p>
								<p class="r8-sub">{booking.timezone} · {booking.duration} min</p>
							</section>
							<section class="r8-panel">
								<p class="r8-head"><Icon icon={usersIcon} width="16" height="16" />Who</p>
								<ul class="r8-people">
									<li>
										<Avatar name={booking.host.name} email={booking.host.email} size={30} rounded="full" />
										<div><p class="r8-name">{booking.host.name}</p><p class="r8-role">Host</p></div>
									</li>
									<li>
										<Avatar name={booking.name} email={booking.email} size={30} rounded="full" />
										<div><p class="r8-name">{booking.name}</p><p class="r8-role">{booking.email}</p></div>
									</li>
									{#each booking.guests as guest (guest)}
										<li>
											<Avatar name={guest} email={guest} size={30} rounded="full" />
											<div><p class="r8-name truncate">{guest}</p><p class="r8-role">Guest</p></div>
										</li>
									{/each}
								</ul>
							</section>
							<section class="r8-panel r8-full">
								<p class="r8-head"><Icon icon={notesIcon} width="16" height="16" />Notes</p>
								<p class="r8-notes">{booking.notes}</p>
							</section>
						</div>
						<div class="mt-6"><Button>Confirm booking</Button></div>

						<!-- ============================== 9 ============================== -->
					{:else if v.n === 9}
						<h3 class="text-base font-semibold">Review your booking</h3>
						<div class="r9">
							<p class="r9-line r9-strong">
								<Icon icon={calendarIcon} width="16" height="16" class="r9-icon" />
								{booking.dateShort} · {booking.range} · {booking.timezone}
							</p>
							<p class="r9-line">
								<Icon icon={userIcon} width="16" height="16" class="r9-icon" />
								{booking.name} · {booking.email}
							</p>
							<p class="r9-line">
								<Icon icon={usersIcon} width="16" height="16" class="r9-icon" />
								<span class="r9-guests">
									{#each booking.guests as guest (guest)}
										<span class="r9-chip">{guest}</span>
									{/each}
								</span>
							</p>
							<p class="r9-line">
								<Icon icon={notesIcon} width="16" height="16" class="r9-icon" />
								<span class="r9-notes">{booking.notes}</span>
							</p>
						</div>
						<div class="mt-5"><Button>Confirm booking</Button></div>

						<!-- ============================== 10 ============================== -->
					{:else if v.n === 10}
						<div class="r10-hero">
							<div class="r10-cal">
								<span class="r10-month">{booking.month}</span>
								<span class="r10-day">{booking.day}</span>
							</div>
							<div>
								<p class="r10-time">{booking.range}</p>
								<p class="r10-sub">{booking.weekday} · {booking.timezone}</p>
								<p class="r10-sub">{booking.event} · {booking.duration} min · {booking.host.name}</p>
							</div>
						</div>
						<dl class="r10-meta">
							<div><dt>You</dt><dd>{booking.name} <span class="r10-dim">{booking.email}</span></dd></div>
							<div>
								<dt>Guests · {guestCount}</dt>
								<dd>
									<ul class="r10-guests">
										{#each booking.guests as guest (guest)}
											<li>{guest}</li>
										{/each}
									</ul>
								</dd>
							</div>
							<div><dt>Notes</dt><dd>{booking.notes}</dd></div>
						</dl>
						<div class="mt-6 r10-cta"><Button>Confirm booking</Button></div>
					{/if}
				</div>
			</article>
		{/each}
	</div>
</main>

<style>
	/* ---------- sandbox chrome (same as the guest-block sandbox) ---------- */
	.page {
		max-width: 68rem;
		margin: 0 auto;
		padding: 2rem 1.5rem 5rem;
		color: rgb(var(--color-text));
	}

	.page-head {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
		margin-bottom: 2rem;
	}

	.card {
		border: 1px solid rgb(var(--color-border));
		border-radius: 1rem;
		background: rgb(var(--color-foreground));
		box-shadow: var(--shadow-small);
		padding: 1.25rem 1.5rem 1.5rem;
	}

	.card-head {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		padding-bottom: 1rem;
		border-bottom: 1px solid rgb(var(--color-border));
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
	}

	/* Each variant is boxed to the real step's column width. */
	.frame {
		max-width: 36rem;
		padding-top: 1.25rem;
	}

	.frame.is-wide {
		max-width: 46rem;
	}

	/* ---------- 1: icon · label · value ---------- */
	.r1-row {
		display: grid;
		grid-template-columns: 1.25rem 6.5rem 1fr;
		gap: 0.875rem;
		align-items: start;
	}

	.r1-row :global(.r1-icon) {
		margin-top: 0.125rem;
		color: rgb(var(--color-primary));
	}

	.r1-label {
		font-size: 0.875rem;
		line-height: 1.5rem;
		color: rgb(var(--color-text) / 0.65);
	}

	.r1-value {
		font-size: 1rem;
		font-weight: 500;
		line-height: 1.5rem;
	}

	.r1-count {
		display: inline-grid;
		place-items: center;
		min-width: 1.25rem;
		padding: 0 0.3125rem;
		border-radius: 999px;
		background: rgb(var(--color-primary) / 0.12);
		color: rgb(var(--color-primary));
		font-size: 0.75rem;
		font-weight: 600;
		line-height: 1.25rem;
	}

	.r1-guests {
		display: grid;
		gap: 0.125rem;
		list-style: none;
		margin: 0;
		padding: 0;
		overflow-wrap: anywhere;
	}

	/* ---------- 2: definition list ---------- */
	.r2 {
		margin: 1.25rem 0 0;
	}

	.r2-row {
		display: grid;
		grid-template-columns: 1fr auto;
		gap: 1.5rem;
		align-items: baseline;
		padding: 0.75rem 0;
		border-bottom: 1px solid rgb(var(--color-border) / 0.7);
	}

	.r2-row:first-child {
		border-top: 1px solid rgb(var(--color-border) / 0.7);
	}

	.r2-row dt {
		font-size: 0.6875rem;
		font-weight: 700;
		letter-spacing: 0.07em;
		text-transform: uppercase;
		color: rgb(var(--color-text) / 0.55);
	}

	.r2-row dd {
		margin: 0;
		font-size: 0.9375rem;
		font-weight: 500;
		text-align: right;
		overflow-wrap: anywhere;
	}

	.r2-guests {
		display: grid;
		gap: 0.125rem;
		list-style: none;
		margin: 0;
		padding: 0;
	}

	/* ---------- 3: ticket ---------- */
	.r3 {
		overflow: hidden;
		border: 1px solid rgb(var(--color-border));
		border-radius: 1rem;
		background: rgb(var(--color-foreground));
	}

	.r3-head {
		background: rgb(var(--color-primary));
		color: rgb(var(--color-contrast-text));
		padding: 1.25rem 1.5rem 1.5rem;
	}

	.r3-event {
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 0.07em;
		text-transform: uppercase;
		opacity: 0.85;
	}

	.r3-time {
		margin-top: 0.5rem;
		font-size: 1.625rem;
		font-weight: 600;
		line-height: 1.15;
	}

	.r3-date {
		margin-top: 0.25rem;
		font-size: 0.9375rem;
		font-weight: 500;
	}

	.r3-tz {
		display: flex;
		align-items: center;
		gap: 0.375rem;
		margin-top: 0.75rem;
		font-size: 0.8125rem;
		opacity: 0.9;
	}

	/* Tear line: dashes across, with a notch bitten out of each edge. */
	.r3-perf {
		position: relative;
		height: 0;
		border-top: 2px dashed rgb(var(--color-border));
	}

	.r3-perf::before,
	.r3-perf::after {
		content: '';
		position: absolute;
		top: -0.625rem;
		width: 1.25rem;
		height: 1.25rem;
		border-radius: 999px;
		background: rgb(var(--color-foreground));
		box-shadow: 0 0 0 1px rgb(var(--color-border));
	}

	.r3-perf::before {
		left: -0.75rem;
	}

	.r3-perf::after {
		right: -0.75rem;
	}

	.r3-body {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 1.25rem;
		padding: 1.5rem;
	}

	.r3-full {
		grid-column: 1 / -1;
	}

	.r3-k {
		font-size: 0.6875rem;
		font-weight: 700;
		letter-spacing: 0.07em;
		text-transform: uppercase;
		color: rgb(var(--color-text) / 0.55);
	}

	.r3-v {
		margin-top: 0.25rem;
		font-size: 0.9375rem;
		font-weight: 500;
		overflow-wrap: anywhere;
	}

	.r3-sub {
		font-size: 0.8125rem;
		color: rgb(var(--color-text) / 0.6);
		overflow-wrap: anywhere;
	}

	.r3-guests {
		display: grid;
		gap: 0.25rem;
		margin: 0.375rem 0 0;
		padding: 0;
		list-style: none;
		font-size: 0.9375rem;
		font-weight: 500;
		overflow-wrap: anywhere;
	}

	.r3-notes {
		font-weight: 400;
		color: rgb(var(--color-text) / 0.85);
	}

	/* ---------- 4: fact cards ---------- */
	.r4 {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 0.75rem;
		margin-top: 1.25rem;
	}

	.r4-card {
		border-radius: 0.875rem;
		background: rgb(var(--color-text) / 0.04);
		padding: 0.875rem 1rem;
	}

	.r4-full {
		grid-column: 1 / -1;
	}

	.r4-k {
		display: flex;
		align-items: center;
		gap: 0.375rem;
		font-size: 0.75rem;
		font-weight: 600;
		color: rgb(var(--color-primary));
	}

	.r4-v {
		margin-top: 0.375rem;
		font-size: 0.9375rem;
		font-weight: 500;
		overflow-wrap: anywhere;
	}

	.r4-sub {
		font-size: 0.8125rem;
		color: rgb(var(--color-text) / 0.6);
		overflow-wrap: anywhere;
	}

	.r4-guests {
		display: grid;
		gap: 0.5rem;
		margin: 0.625rem 0 0;
		padding: 0;
		list-style: none;
	}

	.r4-guests li {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.875rem;
		font-weight: 500;
	}

	.r4-notes {
		font-weight: 400;
	}

	/* ---------- 5: timeline rail ---------- */
	.r5 {
		margin: 1.25rem 0 0;
		padding: 0;
		list-style: none;
	}

	.r5-item {
		position: relative;
		display: grid;
		grid-template-columns: 1.75rem 1fr;
		gap: 0.875rem;
		padding-bottom: 1.25rem;
	}

	/* The rail runs from each bubble down to the next one. */
	.r5-item::before {
		content: '';
		position: absolute;
		top: 1.75rem;
		bottom: 0;
		left: 0.8125rem;
		width: 2px;
		background: rgb(var(--color-primary) / 0.2);
	}

	.r5-last {
		padding-bottom: 0;
	}

	.r5-last::before {
		display: none;
	}

	.r5-bubble {
		display: grid;
		place-items: center;
		width: 1.75rem;
		height: 1.75rem;
		border-radius: 999px;
		background: rgb(var(--color-primary) / 0.14);
		color: rgb(var(--color-primary));
	}

	.r5-k {
		font-size: 0.75rem;
		font-weight: 600;
		color: rgb(var(--color-text) / 0.55);
	}

	.r5-v {
		font-size: 0.9375rem;
		font-weight: 500;
		overflow-wrap: anywhere;
	}

	.r5-sub {
		font-size: 0.8125rem;
		color: rgb(var(--color-text) / 0.6);
	}

	.r5-guests {
		display: grid;
		gap: 0.125rem;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.r5-notes {
		font-weight: 400;
	}

	/* ---------- 6: sentence summary ---------- */
	.r6-lede {
		font-size: 1.125rem;
		line-height: 1.5;
	}

	.r6-lede strong {
		font-weight: 600;
		color: rgb(var(--color-primary));
	}

	.r6-meta {
		margin-top: 0.625rem;
		font-size: 0.8125rem;
		color: rgb(var(--color-text) / 0.6);
	}

	.r6-block {
		margin-top: 1.5rem;
	}

	.r6-k {
		font-size: 0.75rem;
		font-weight: 600;
		color: rgb(var(--color-text) / 0.55);
	}

	.r6-chips {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		margin: 0.625rem 0 0;
		padding: 0;
		list-style: none;
	}

	.r6-chip {
		display: inline-flex;
		align-items: center;
		gap: 0.4rem;
		max-width: 100%;
		border: 1px solid rgb(var(--color-border));
		border-radius: 999px;
		padding: 0.25rem 0.75rem 0.25rem 0.25rem;
		font-size: 0.8125rem;
		font-weight: 600;
	}

	.r6-quote {
		margin-top: 1.5rem;
		border-left: 3px solid rgb(var(--color-primary) / 0.4);
		padding: 0.125rem 0 0.125rem 0.875rem;
		font-size: 0.9375rem;
		color: rgb(var(--color-text) / 0.8);
	}

	/* ---------- 7: editable rows ---------- */
	.r7 {
		margin: 1.25rem 0 0;
		padding: 0;
		list-style: none;
		border-radius: 0.875rem;
		overflow: hidden;
	}

	.r7-row {
		display: grid;
		grid-template-columns: 1.125rem 1fr auto;
		gap: 0.75rem;
		align-items: start;
		padding: 0.875rem 1rem;
	}

	.r7-row:nth-child(odd) {
		background: rgb(var(--color-text) / 0.04);
	}

	.r7-row :global(.r7-icon) {
		margin-top: 0.1875rem;
		color: rgb(var(--color-primary));
	}

	.r7-k {
		font-size: 0.75rem;
		color: rgb(var(--color-text) / 0.55);
	}

	.r7-v {
		font-size: 0.9375rem;
		font-weight: 500;
		overflow-wrap: anywhere;
	}

	.r7-guests {
		display: grid;
		gap: 0.125rem;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.r7-edit {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		border: 0;
		background: transparent;
		padding: 0.125rem 0;
		color: rgb(var(--color-primary));
		font-size: 0.8125rem;
		font-weight: 600;
		cursor: pointer;
	}

	.r7-edit:hover {
		text-decoration: underline;
	}

	/* ---------- 8: when / who split ---------- */
	.r8 {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 0.75rem;
		margin-top: 1.25rem;
	}

	.r8-panel {
		border: 1px solid rgb(var(--color-border));
		border-radius: 0.875rem;
		padding: 1rem 1.125rem 1.125rem;
	}

	.r8-full {
		grid-column: 1 / -1;
	}

	.r8-head {
		display: flex;
		align-items: center;
		gap: 0.375rem;
		padding-bottom: 0.75rem;
		border-bottom: 1px solid rgb(var(--color-border) / 0.7);
		color: rgb(var(--color-primary));
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 0.06em;
		text-transform: uppercase;
	}

	.r8-day {
		margin-top: 0.875rem;
		font-size: 0.8125rem;
		color: rgb(var(--color-text) / 0.6);
	}

	.r8-date {
		font-size: 1.25rem;
		font-weight: 600;
	}

	.r8-time {
		margin-top: 0.5rem;
		font-size: 1.0625rem;
		font-weight: 500;
	}

	.r8-sub {
		margin-top: 0.125rem;
		font-size: 0.8125rem;
		color: rgb(var(--color-text) / 0.6);
	}

	.r8-people {
		display: grid;
		gap: 0.625rem;
		margin: 0.875rem 0 0;
		padding: 0;
		list-style: none;
	}

	.r8-people li {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr);
		gap: 0.625rem;
		align-items: center;
	}

	.r8-name {
		font-size: 0.875rem;
		font-weight: 500;
	}

	.r8-role {
		font-size: 0.75rem;
		color: rgb(var(--color-text) / 0.55);
	}

	.r8-notes {
		margin-top: 0.875rem;
		font-size: 0.9375rem;
	}

	/* ---------- 9: dense meta lines ---------- */
	.r9 {
		display: grid;
		gap: 0.5rem;
		margin-top: 0.875rem;
	}

	.r9-line {
		display: grid;
		grid-template-columns: 1rem 1fr;
		gap: 0.625rem;
		align-items: start;
		font-size: 0.875rem;
		overflow-wrap: anywhere;
	}

	.r9-line :global(.r9-icon) {
		margin-top: 0.1875rem;
		color: rgb(var(--color-primary));
	}

	.r9-strong {
		font-size: 0.9375rem;
		font-weight: 600;
	}

	.r9-guests {
		display: flex;
		flex-wrap: wrap;
		gap: 0.25rem;
	}

	.r9-chip {
		border-radius: 0.375rem;
		background: rgb(var(--color-primary) / 0.1);
		color: rgb(var(--color-primary));
		padding: 0.0625rem 0.375rem;
		font-size: 0.8125rem;
		font-weight: 500;
	}

	.r9-notes {
		color: rgb(var(--color-text) / 0.75);
	}

	/* ---------- 10: hero date ---------- */
	.r10-hero {
		display: flex;
		align-items: center;
		gap: 1.25rem;
	}

	.r10-cal {
		display: grid;
		place-items: center;
		flex-shrink: 0;
		width: 5rem;
		border-radius: 1rem;
		background: rgb(var(--color-primary) / 0.12);
		padding: 0.5rem 0 0.75rem;
		color: rgb(var(--color-primary));
	}

	.r10-month {
		font-size: 0.75rem;
		font-weight: 700;
		letter-spacing: 0.08em;
		text-transform: uppercase;
	}

	.r10-day {
		font-size: 2.5rem;
		font-weight: 600;
		line-height: 1.05;
	}

	.r10-time {
		font-size: 1.375rem;
		font-weight: 600;
	}

	.r10-sub {
		font-size: 0.8125rem;
		color: rgb(var(--color-text) / 0.6);
	}

	.r10-meta {
		display: grid;
		gap: 0.875rem;
		margin: 1.5rem 0 0;
	}

	.r10-meta dt {
		font-size: 0.75rem;
		color: rgb(var(--color-text) / 0.55);
	}

	.r10-meta dd {
		margin: 0.125rem 0 0;
		font-size: 0.9375rem;
		font-weight: 500;
		overflow-wrap: anywhere;
	}

	.r10-dim {
		font-weight: 400;
		color: rgb(var(--color-text) / 0.6);
	}

	.r10-guests {
		display: grid;
		gap: 0.125rem;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	/* Full-width confirm reads as the end of the block. */
	.r10-cta :global(button) {
		width: 100%;
		justify-content: center;
	}

	@media (max-width: 640px) {
		.r3-body,
		.r4,
		.r8 {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
