<script lang="ts">
	// Design sandbox: ten looks for the booking form's guest block, each shown in all
	// three states (empty / typing a guest / guests added). Pick one and it gets wired
	// into GuestEmailFields.svelte.
	import Icon from '@iconify/svelte';
	import plusIcon from '@iconify-icons/tabler/plus';
	import checkIcon from '@iconify-icons/tabler/check';
	import xIcon from '@iconify-icons/tabler/x';
	import trashIcon from '@iconify-icons/tabler/trash';
	import mailIcon from '@iconify-icons/tabler/mail';
	import userPlusIcon from '@iconify-icons/tabler/user-plus';
	import usersIcon from '@iconify-icons/tabler/users';
	import Avatar from '$lib/components/ui/Avatar.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import IconButton from '$lib/components/ui/IconButton.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';

	const bold = (icon: { body: string }) => ({
		...icon,
		body: icon.body.replace('stroke-width="2"', 'stroke-width="3"')
	});
	const boldPlus = bold(plusIcon);
	const boldCheck = bold(checkIcon);
	const boldX = bold(xIcon);

	const guests = ['nina@acme.io', 'marcus.long.name@company-with-a-long-domain.com'];
	const draft = 'nina@acme.io';

	const variants = [
		{ n: 1, title: 'Pill chips + tinted squares', note: 'Current build: rounded chips, cross on hover, 48px tinted action squares.' },
		{ n: 2, title: 'Tag field', note: 'Chips live inside one bordered field; you type at the end of the list.' },
		{ n: 3, title: 'Guest rows', note: 'Each guest is a card row with avatar and trash button — closest to the users table.' },
		{ n: 4, title: 'Actions inside the field', note: 'Check/cross sit inside the input on the right, so the row is one object.' },
		{ n: 5, title: 'Text buttons', note: 'No icon buttons at all — Add / Cancel as a primary + ghost text pair.' },
		{ n: 6, title: 'Avatar chips + add chip', note: 'Chips carry initials and a permanent cross; a dashed chip opens the field.' },
		{ n: 7, title: 'Dashed drop zone', note: 'Empty state is a dashed panel that expands into the field in place.' },
		{ n: 8, title: 'Counted list', note: 'Header with count and a right-aligned ghost link; guests separated by dividers.' },
		{ n: 9, title: 'Solid brand chips', note: 'Filled brand chips, full-width soft add button, remaining-slots hint.' },
		{ n: 10, title: 'Label-left form row', note: 'Field label sits in a left column, matching a dense settings layout.' }
	];
</script>

<svelte:head><title>Guest block variants</title></svelte:head>

<main class="page">
	<header class="page-head">
		<div>
			<h1 class="text-2xl font-semibold">Guest block — 10 variants</h1>
			<p class="mt-2 text-sm opacity-70">Each card shows the same three states. Tell me a number and I'll make it the real one.</p>
		</div>
		<ThemeToggle />
	</header>

	<div class="grid gap-6">
		{#each variants as v (v.n)}
			<article class="card">
				<div class="card-head">
					<span class="badge">{v.n}</span>
					<div>
						<h2 class="text-base font-semibold">{v.title}</h2>
						<p class="mt-1 text-xs opacity-65">{v.note}</p>
					</div>
				</div>

				<div class="states">
					<!-- ============================== 1 ============================== -->
					{#if v.n === 1}
						<section class="state"><p class="state-label">Empty</p>
							<Button class="gap-2"><Icon icon={boldPlus} width="20" height="20" />Add Guests</Button>
						</section>
						<section class="state"><p class="state-label">Adding</p>
							<div class="grid gap-2 sm:grid-cols-[1fr_auto] sm:items-start sm:gap-3 md:w-1/2">
								<Input id="v1-draft" label="Guest email" type="email" value={draft} />
								<div class="v1-actions flex gap-2">
									<IconButton tone="primary" label="Confirm"><Icon icon={boldCheck} width="20" height="20" /></IconButton>
									<IconButton tone="danger" label="Discard"><Icon icon={boldX} width="20" height="20" /></IconButton>
								</div>
							</div>
						</section>
						<section class="state"><p class="state-label">With guests</p>
							<ul class="flex flex-wrap gap-2">
								{#each guests as g (g)}
									<li class="chip-pill"><span class="truncate">{g}</span>
										<button type="button" class="chip-x" aria-label="Remove"><Icon icon={boldX} width="14" height="14" /></button>
									</li>
								{/each}
							</ul>
							<div class="mt-3"><Button class="gap-2"><Icon icon={boldPlus} width="20" height="20" />Add Guests</Button></div>
						</section>

					<!-- ============================== 2 ============================== -->
					{:else if v.n === 2}
						<section class="state"><p class="state-label">Empty</p>
							<div class="tag-field md:w-1/2"><span class="tag-hint">Add guest emails…</span></div>
						</section>
						<section class="state"><p class="state-label">Adding</p>
							<div class="tag-field is-focus md:w-1/2">
								<span class="tag">{guests[0]}<button type="button" class="tag-x" aria-label="Remove"><Icon icon={boldX} width="12" height="12" /></button></span>
								<span class="tag-caret">{draft}<i class="caret"></i></span>
							</div>
							<p class="helper">Press Enter to add</p>
						</section>
						<section class="state"><p class="state-label">With guests</p>
							<div class="tag-field md:w-1/2">
								{#each guests as g (g)}
									<span class="tag">{g}<button type="button" class="tag-x" aria-label="Remove"><Icon icon={boldX} width="12" height="12" /></button></span>
								{/each}
								<span class="tag-hint">Add another…</span>
							</div>
						</section>

					<!-- ============================== 3 ============================== -->
					{:else if v.n === 3}
						<section class="state"><p class="state-label">Empty</p>
							<Button variant="secondary" class="gap-2"><Icon icon={userPlusIcon} width="20" height="20" />Add guest</Button>
						</section>
						<section class="state"><p class="state-label">Adding</p>
							<div class="row-card md:w-2/3">
								<div class="grow"><Input id="v3-draft" label="Guest email" type="email" value={draft} /></div>
								<IconButton tone="primary" label="Confirm"><Icon icon={boldCheck} width="20" height="20" /></IconButton>
								<IconButton tone="danger" label="Discard"><Icon icon={boldX} width="20" height="20" /></IconButton>
							</div>
						</section>
						<section class="state"><p class="state-label">With guests</p>
							<ul class="grid gap-2 md:w-2/3">
								{#each guests as g (g)}
									<li class="row-card">
										<Avatar email={g} size={34} />
										<span class="grow truncate text-sm font-medium">{g}</span>
										<IconButton tone="danger" label="Remove"><Icon icon={trashIcon} width="18" height="18" /></IconButton>
									</li>
								{/each}
							</ul>
							<div class="mt-3"><Button variant="secondary" class="gap-2"><Icon icon={userPlusIcon} width="20" height="20" />Add guest</Button></div>
						</section>

					<!-- ============================== 4 ============================== -->
					{:else if v.n === 4}
						<section class="state"><p class="state-label">Empty</p>
							<Button class="gap-2"><Icon icon={boldPlus} width="20" height="20" />Add Guests</Button>
						</section>
						<section class="state"><p class="state-label">Adding</p>
							<div class="inset-field md:w-1/2">
								<Icon icon={mailIcon} width="18" height="18" class="inset-icon" />
								<span class="inset-value">{draft}</span>
								<span class="inset-label">Guest email</span>
								<span class="inset-actions">
									<button type="button" class="inset-btn ok" aria-label="Confirm"><Icon icon={boldCheck} width="16" height="16" /></button>
									<button type="button" class="inset-btn no" aria-label="Discard"><Icon icon={boldX} width="16" height="16" /></button>
								</span>
							</div>
						</section>
						<section class="state"><p class="state-label">With guests</p>
							<ul class="flex flex-wrap gap-2">
								{#each guests as g (g)}
									<li class="chip-pill"><span class="truncate">{g}</span><button type="button" class="chip-x" aria-label="Remove"><Icon icon={boldX} width="14" height="14" /></button></li>
								{/each}
							</ul>
							<div class="mt-3"><Button class="gap-2"><Icon icon={boldPlus} width="20" height="20" />Add Guests</Button></div>
						</section>

					<!-- ============================== 5 ============================== -->
					{:else if v.n === 5}
						<section class="state"><p class="state-label">Empty</p>
							<Button variant="soft" class="gap-2"><Icon icon={boldPlus} width="20" height="20" />Add Guests</Button>
						</section>
						<section class="state"><p class="state-label">Adding</p>
							<div class="grid gap-3 md:w-1/2">
								<Input id="v5-draft" label="Guest email" type="email" value={draft} />
								<div class="flex gap-2">
									<Button size="small">Add guest</Button>
									<Button variant="ghost" size="small">Cancel</Button>
								</div>
							</div>
						</section>
						<section class="state"><p class="state-label">With guests</p>
							<ul class="flex flex-wrap gap-2">
								{#each guests as g (g)}
									<li class="chip-pill"><span class="truncate">{g}</span><button type="button" class="chip-x" aria-label="Remove"><Icon icon={boldX} width="14" height="14" /></button></li>
								{/each}
							</ul>
							<div class="mt-3"><Button variant="soft" class="gap-2"><Icon icon={boldPlus} width="20" height="20" />Add Guests</Button></div>
						</section>

					<!-- ============================== 6 ============================== -->
					{:else if v.n === 6}
						<section class="state"><p class="state-label">Empty</p>
							<button type="button" class="chip-add"><Icon icon={boldPlus} width="16" height="16" />Add guest</button>
						</section>
						<section class="state"><p class="state-label">Adding</p>
							<div class="grid gap-2 sm:grid-cols-[1fr_auto] sm:items-start sm:gap-3 md:w-1/2">
								<Input id="v6-draft" label="Guest email" type="email" value={draft} />
								<div class="v6-actions flex gap-2">
									<button type="button" class="round ok" aria-label="Confirm"><Icon icon={boldCheck} width="18" height="18" /></button>
									<button type="button" class="round no" aria-label="Discard"><Icon icon={boldX} width="18" height="18" /></button>
								</div>
							</div>
						</section>
						<section class="state"><p class="state-label">With guests</p>
							<ul class="flex flex-wrap items-center gap-2">
								{#each guests as g (g)}
									<li class="chip-avatar">
										<Avatar email={g} size={22} />
										<span class="truncate">{g}</span>
										<button type="button" class="chip-x always" aria-label="Remove"><Icon icon={boldX} width="14" height="14" /></button>
									</li>
								{/each}
								<li><button type="button" class="chip-add"><Icon icon={boldPlus} width="16" height="16" />Add guest</button></li>
							</ul>
						</section>

					<!-- ============================== 7 ============================== -->
					{:else if v.n === 7}
						<section class="state"><p class="state-label">Empty</p>
							<button type="button" class="dashed md:w-1/2">
								<Icon icon={usersIcon} width="22" height="22" />
								<span><b>Add guests</b><br /><span class="text-xs opacity-65">Invite others to this meeting</span></span>
							</button>
						</section>
						<section class="state"><p class="state-label">Adding</p>
							<div class="dashed is-open md:w-1/2">
								<div class="grid gap-2 sm:grid-cols-[1fr_auto] sm:items-start sm:gap-3">
									<Input id="v7-draft" label="Guest email" type="email" value={draft} />
									<div class="v1-actions flex gap-2">
										<IconButton tone="primary" label="Confirm"><Icon icon={boldCheck} width="20" height="20" /></IconButton>
										<IconButton tone="danger" label="Discard"><Icon icon={boldX} width="20" height="20" /></IconButton>
									</div>
								</div>
							</div>
						</section>
						<section class="state"><p class="state-label">With guests</p>
							<div class="dashed is-open md:w-1/2">
								<ul class="flex flex-wrap gap-2">
									{#each guests as g (g)}
										<li class="chip-pill"><span class="truncate">{g}</span><button type="button" class="chip-x" aria-label="Remove"><Icon icon={boldX} width="14" height="14" /></button></li>
									{/each}
								</ul>
								<button type="button" class="link-add mt-3"><Icon icon={boldPlus} width="16" height="16" />Add another guest</button>
							</div>
						</section>

					<!-- ============================== 8 ============================== -->
					{:else if v.n === 8}
						<section class="state"><p class="state-label">Empty</p>
							<div class="md:w-2/3">
								<div class="count-head"><span>Guests <span class="opacity-60">(0)</span></span>
									<button type="button" class="link-add"><Icon icon={boldPlus} width="16" height="16" />Add guest</button>
								</div>
								<p class="helper">No guests yet.</p>
							</div>
						</section>
						<section class="state"><p class="state-label">Adding</p>
							<div class="md:w-2/3">
								<div class="count-head"><span>Guests <span class="opacity-60">(1)</span></span></div>
								<div class="mt-3 grid gap-2 sm:grid-cols-[1fr_auto] sm:items-start sm:gap-3">
									<Input id="v8-draft" label="Guest email" type="email" value={draft} />
									<div class="v1-actions flex gap-2">
										<IconButton tone="primary" label="Confirm"><Icon icon={boldCheck} width="20" height="20" /></IconButton>
										<IconButton tone="danger" label="Discard"><Icon icon={boldX} width="20" height="20" /></IconButton>
									</div>
								</div>
							</div>
						</section>
						<section class="state"><p class="state-label">With guests</p>
							<div class="md:w-2/3">
								<div class="count-head"><span>Guests <span class="opacity-60">({guests.length})</span></span>
									<button type="button" class="link-add"><Icon icon={boldPlus} width="16" height="16" />Add guest</button>
								</div>
								<ul class="divided">
									{#each guests as g (g)}
										<li><span class="truncate text-sm">{g}</span><button type="button" class="ghost-x" aria-label="Remove"><Icon icon={boldX} width="16" height="16" /></button></li>
									{/each}
								</ul>
							</div>
						</section>

					<!-- ============================== 9 ============================== -->
					{:else if v.n === 9}
						<section class="state"><p class="state-label">Empty</p>
							<div class="md:w-1/2"><Button variant="soft" fullWidth class="gap-2"><Icon icon={boldPlus} width="20" height="20" />Add Guests</Button></div>
						</section>
						<section class="state"><p class="state-label">Adding</p>
							<div class="grid gap-2 sm:grid-cols-[1fr_auto] sm:items-start sm:gap-3 md:w-1/2">
								<Input id="v9-draft" label="Guest email" type="email" value={draft} />
								<div class="v9-actions flex gap-2">
									<button type="button" class="solid ok" aria-label="Confirm"><Icon icon={boldCheck} width="20" height="20" /></button>
									<button type="button" class="solid no" aria-label="Discard"><Icon icon={boldX} width="20" height="20" /></button>
								</div>
							</div>
						</section>
						<section class="state"><p class="state-label">With guests</p>
							<ul class="flex flex-wrap gap-2">
								{#each guests as g (g)}
									<li class="chip-solid"><span class="truncate">{g}</span><button type="button" class="chip-x always solid-x" aria-label="Remove"><Icon icon={boldX} width="14" height="14" /></button></li>
								{/each}
							</ul>
							<p class="helper">2 of 5 guest slots used</p>
							<div class="mt-3 md:w-1/2"><Button variant="soft" fullWidth class="gap-2"><Icon icon={boldPlus} width="20" height="20" />Add Guests</Button></div>
						</section>

					<!-- ============================== 10 ============================= -->
					{:else}
						<section class="state"><p class="state-label">Empty</p>
							<div class="form-row"><span class="form-label">Guests</span>
								<Button variant="secondary" size="small" class="gap-2"><Icon icon={boldPlus} width="18" height="18" />Add</Button>
							</div>
						</section>
						<section class="state"><p class="state-label">Adding</p>
							<div class="form-row"><span class="form-label">Guests</span>
								<div class="grid grow gap-2 sm:grid-cols-[1fr_auto] sm:items-start sm:gap-3">
									<Input id="v10-draft" label="Guest email" type="email" value={draft} />
									<div class="v1-actions flex gap-2">
										<IconButton tone="primary" label="Confirm"><Icon icon={boldCheck} width="20" height="20" /></IconButton>
										<IconButton tone="danger" label="Discard"><Icon icon={boldX} width="20" height="20" /></IconButton>
									</div>
								</div>
							</div>
						</section>
						<section class="state"><p class="state-label">With guests</p>
							<div class="form-row"><span class="form-label">Guests</span>
								<div class="grow">
									<ul class="flex flex-wrap gap-2">
										{#each guests as g (g)}
											<li class="chip-pill"><span class="truncate">{g}</span><button type="button" class="chip-x" aria-label="Remove"><Icon icon={boldX} width="14" height="14" /></button></li>
										{/each}
									</ul>
									<button type="button" class="link-add mt-3"><Icon icon={boldPlus} width="16" height="16" />Add guest</button>
								</div>
							</div>
						</section>
					{/if}
				</div>
			</article>
		{/each}
	</div>
</main>

<style>
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

	.states {
		display: grid;
		gap: 1.5rem;
		padding-top: 1.25rem;
	}

	.state-label {
		margin-bottom: 0.625rem;
		font-size: 0.6875rem;
		font-weight: 700;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		opacity: 0.5;
	}

	.helper {
		margin-top: 0.5rem;
		font-size: 0.75rem;
		opacity: 0.65;
	}

	/* ---------- shared chips ---------- */
	.chip-pill,
	.chip-avatar,
	.chip-solid {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		max-width: 100%;
		border-radius: 999px;
		font-size: 0.8125rem;
		font-weight: 600;
		line-height: 1.25;
	}

	.chip-pill {
		border: 1px solid rgb(var(--color-border));
		background: rgb(var(--color-text) / 0.05);
		padding: 0.3rem 0.5rem 0.3rem 0.75rem;
	}

	.chip-avatar {
		border: 1px solid rgb(var(--color-border));
		background: rgb(var(--color-foreground));
		padding: 0.2rem 0.4rem 0.2rem 0.25rem;
		gap: 0.4rem;
	}

	.chip-solid {
		background: rgb(var(--color-primary) / 0.16);
		color: rgb(var(--color-primary));
		padding: 0.35rem 0.5rem 0.35rem 0.8rem;
	}

	.chip-x {
		display: grid;
		place-items: center;
		flex-shrink: 0;
		width: 1.125rem;
		height: 1.125rem;
		border-radius: 999px;
		cursor: pointer;
		opacity: 0;
		color: rgb(var(--error));
		transition: opacity 0.15s ease, background-color 0.15s ease;
	}

	.chip-x.always {
		opacity: 0.55;
	}

	.chip-x.solid-x {
		color: rgb(var(--color-primary));
	}

	.chip-pill:hover .chip-x,
	.chip-avatar:hover .chip-x,
	.chip-solid:hover .chip-x,
	.chip-x:focus-visible {
		opacity: 1;
	}

	.chip-x:hover {
		background: rgb(var(--error) / 0.16);
	}

	.chip-add {
		display: inline-flex;
		align-items: center;
		gap: 0.3rem;
		border: 1px dashed rgb(var(--color-primary) / 0.55);
		border-radius: 999px;
		padding: 0.35rem 0.75rem;
		color: rgb(var(--color-primary));
		background: transparent;
		font-size: 0.8125rem;
		font-weight: 600;
		cursor: pointer;
		transition: background-color 0.15s ease;
	}

	.chip-add:hover {
		background: rgb(var(--color-primary) / 0.1);
	}

	.link-add {
		display: inline-flex;
		align-items: center;
		gap: 0.3rem;
		background: transparent;
		border: 0;
		padding: 0;
		color: rgb(var(--color-primary));
		font-size: 0.8125rem;
		font-weight: 600;
		cursor: pointer;
	}

	.link-add:hover {
		text-decoration: underline;
	}

	/* ---------- 1 / 7 / 8 / 10: tinted squares ---------- */
	.v1-actions :global(.icon-button) {
		width: 48px;
		height: 48px;
		border-radius: 11px;
	}

	.v1-actions :global(.tone-primary) {
		background: rgb(var(--color-primary) / 0.14);
		color: rgb(var(--color-primary));
	}

	.v1-actions :global(.tone-danger) {
		background: rgb(var(--error) / 0.14);
		color: rgb(var(--error));
	}

	/* ---------- 2: tag field ---------- */
	.tag-field {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.375rem;
		min-height: 48px;
		border: 2px solid rgb(var(--color-border));
		border-radius: 10px;
		background: rgb(var(--color-foreground));
		padding: 0.5rem 0.625rem;
	}

	.tag-field.is-focus {
		border-color: rgb(var(--color-primary));
		box-shadow: 0 0 0 3px rgb(var(--color-primary) / 0.25);
	}

	.tag {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
		max-width: 100%;
		border-radius: 6px;
		background: rgb(var(--color-primary) / 0.12);
		color: rgb(var(--color-primary));
		padding: 0.2rem 0.35rem 0.2rem 0.5rem;
		font-size: 0.8125rem;
		font-weight: 600;
	}

	.tag-x {
		display: grid;
		place-items: center;
		width: 1rem;
		height: 1rem;
		border-radius: 4px;
		cursor: pointer;
		opacity: 0.7;
	}

	.tag-x:hover {
		opacity: 1;
		background: rgb(var(--color-primary) / 0.18);
	}

	.tag-hint {
		font-size: 0.875rem;
		opacity: 0.45;
	}

	.tag-caret {
		display: inline-flex;
		align-items: center;
		font-size: 0.875rem;
	}

	.caret {
		display: inline-block;
		width: 1px;
		height: 1.05em;
		margin-left: 1px;
		background: rgb(var(--color-primary));
	}

	/* ---------- 3: guest rows ---------- */
	.row-card {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		border: 1px solid rgb(var(--color-border));
		border-radius: 12px;
		background: rgb(var(--color-foreground));
		padding: 0.5rem 0.625rem;
	}

	/* ---------- 4: actions inside the field ---------- */
	.inset-field {
		position: relative;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		min-height: 48px;
		border: 2px solid rgb(var(--color-primary));
		border-radius: 10px;
		background: rgb(var(--color-foreground));
		box-shadow: 0 0 0 3px rgb(var(--color-primary) / 0.25);
		padding: 0 6px 0 12px;
	}

	.inset-field :global(.inset-icon) {
		flex-shrink: 0;
		color: rgb(var(--color-primary));
	}

	.inset-value {
		flex: 1;
		font-size: 0.9rem;
	}

	.inset-label {
		position: absolute;
		top: -0.55rem;
		left: 0.6rem;
		padding: 0 0.25rem;
		background: rgb(var(--color-foreground));
		color: rgb(var(--color-primary));
		font-size: 0.72rem;
	}

	.inset-actions {
		display: flex;
		gap: 4px;
	}

	.inset-btn {
		display: grid;
		place-items: center;
		width: 34px;
		height: 34px;
		border-radius: 8px;
		cursor: pointer;
	}

	.inset-btn.ok {
		color: rgb(var(--color-primary));
	}

	.inset-btn.ok:hover {
		background: rgb(var(--color-primary) / 0.14);
	}

	.inset-btn.no {
		color: rgb(var(--error));
	}

	.inset-btn.no:hover {
		background: rgb(var(--error) / 0.14);
	}

	/* ---------- 6: round actions ---------- */
	.v6-actions {
		height: 44px;
		align-items: center;
	}

	.round {
		display: grid;
		place-items: center;
		width: 38px;
		height: 38px;
		border-radius: 999px;
		cursor: pointer;
	}

	.round.ok {
		background: rgb(var(--color-primary) / 0.14);
		color: rgb(var(--color-primary));
	}

	.round.no {
		background: rgb(var(--error) / 0.14);
		color: rgb(var(--error));
	}

	/* ---------- 7: dashed panel ---------- */
	.dashed {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		width: 100%;
		border: 2px dashed rgb(var(--color-border));
		border-radius: 14px;
		background: transparent;
		padding: 1rem;
		text-align: left;
		font-size: 0.875rem;
		cursor: pointer;
		color: rgb(var(--color-text));
	}

	.dashed:hover {
		border-color: rgb(var(--color-primary) / 0.6);
		background: rgb(var(--color-primary) / 0.05);
	}

	.dashed.is-open {
		display: block;
		cursor: default;
		border-style: solid;
		border-color: rgb(var(--color-border));
	}

	/* ---------- 8: counted list ---------- */
	.count-head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		padding-bottom: 0.5rem;
		border-bottom: 1px solid rgb(var(--color-border));
		font-size: 0.875rem;
		font-weight: 600;
	}

	.divided {
		display: grid;
	}

	.divided li {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem;
		padding: 0.625rem 0;
		border-bottom: 1px solid rgb(var(--color-border) / 0.6);
	}

	.ghost-x {
		display: grid;
		place-items: center;
		width: 28px;
		height: 28px;
		border-radius: 8px;
		cursor: pointer;
		opacity: 0.55;
	}

	.ghost-x:hover {
		opacity: 1;
		background: rgb(var(--error) / 0.14);
		color: rgb(var(--error));
	}

	/* ---------- 9: solid actions ---------- */
	.v9-actions {
		height: 48px;
		align-items: center;
	}

	.solid {
		display: grid;
		place-items: center;
		width: 48px;
		height: 48px;
		border-radius: 11px;
		cursor: pointer;
	}

	.solid.ok {
		background: rgb(var(--color-primary));
		color: rgb(var(--color-contrast-text));
	}

	.solid.no {
		background: rgb(var(--error));
		color: #fff;
	}

	/* ---------- 10: label-left row ---------- */
	.form-row {
		display: flex;
		align-items: flex-start;
		gap: 1rem;
	}

	.form-label {
		flex-shrink: 0;
		width: 6.5rem;
		padding-top: 0.7rem;
		font-size: 0.875rem;
		font-weight: 600;
	}

	@media (prefers-reduced-motion: reduce) {
		.chip-x,
		.chip-add {
			transition: none;
		}
	}
</style>
