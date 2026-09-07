<script lang="ts">
	import Icon from '@iconify/svelte';
	import chevronLeftIcon from '@iconify-icons/tabler/chevron-left';
	import chevronRightIcon from '@iconify-icons/tabler/chevron-right';

	let {
		month = $bindable(),
		selected = $bindable(),
		availableDates,
		minimumMonth,
		today = ''
	}: {
		month: string;
		selected: string;
		availableDates: string[];
		minimumMonth: string;
		today?: string;
	} = $props();

	const year = $derived(Number(month.slice(0, 4)));
	const monthNumber = $derived(Number(month.slice(5, 7)));
	const daysInMonth = $derived(new Date(Date.UTC(year, monthNumber, 0)).getUTCDate());
	const leadingDays = $derived((new Date(Date.UTC(year, monthNumber - 1, 1)).getUTCDay() + 6) % 7);
	// Restrict to the current month so stale availability (from the month just left)
	// can never match this month's date keys during a reactive update.
	const available = $derived(new Set(availableDates.filter((date) => date.startsWith(month))));
	let monthDirection = $state(1);
	const monthLabel = $derived(
		new Intl.DateTimeFormat(undefined, { month: 'long', year: 'numeric', timeZone: 'UTC' }).format(
			new Date(Date.UTC(year, monthNumber - 1, 1))
		)
	);

	function moveMonth(amount: number) {
		monthDirection = amount;
		const next = new Date(Date.UTC(year, monthNumber - 1 + amount, 1));
		month = `${next.getUTCFullYear()}-${String(next.getUTCMonth() + 1).padStart(2, '0')}`;
	}

	function keyFor(day: number) {
		return `${month}-${String(day).padStart(2, '0')}`;
	}

	function cellClass(date: string): string {
		// Segmented-tray look: the grid sits in an inset panel; bookable days are
		// raised "chips", the selection fills primary, unavailable days recede.
		const base = 'relative isolate aspect-square w-full overflow-hidden rounded-[10px] text-sm font-bold transition duration-150';
		if (selected === date)
			return `${base} calendar-selected z-10 bg-[rgb(var(--color-foreground))] font-bold text-[rgb(var(--color-contrast-text))]`;
		if (date === today)
			// Today stands out with a bold primary number (plus the dot marker).
			// Only give it a chip background when it actually has bookable times.
			return available.has(date)
				? `${base} day-cell cursor-pointer bg-[rgb(var(--color-foreground))] font-bold text-[rgb(var(--color-primary))]`
				: `${base} font-bold text-[rgb(var(--color-primary))] cursor-not-allowed`;
		if (available.has(date))
			return `${base} day-cell cursor-pointer bg-[rgb(var(--color-foreground))] text-[rgb(var(--color-text))]`;
		return `${base} text-[rgb(var(--color-text)/0.35)]`;
	}
</script>

<div class="calendar-shell w-full overflow-hidden rounded-xl" aria-label={monthLabel}>
	<div class="flex items-center justify-between gap-3 rounded-t-xl bg-[rgb(var(--color-primary))] px-4 pt-2 pb-0">
		{#key month}
			<h2
				class="calendar-label text-lg font-semibold text-[rgb(var(--color-contrast-text))]"
				class:calendar-label-next={monthDirection > 0}
				class:calendar-label-previous={monthDirection < 0}
			>
				{monthLabel}
			</h2>
		{/key}
		<div class="flex items-center gap-2">
			<button
				type="button"
				class="group grid size-11 cursor-pointer place-items-center rounded-xl bg-transparent text-[rgb(var(--color-contrast-text))] transition-colors hover:bg-[rgb(var(--color-contrast-text)/0.15)] disabled:cursor-not-allowed disabled:opacity-30 disabled:hover:bg-transparent"
				disabled={month <= minimumMonth}
				onclick={() => moveMonth(-1)}
				aria-label="Previous month"
			>
				<span class="grid transition-transform duration-200 group-hover:-translate-x-0.5 group-active:-translate-x-1">
					<Icon icon={chevronLeftIcon} width="20" height="20" />
				</span>
			</button>
			<button
				type="button"
				class="group grid size-11 cursor-pointer place-items-center rounded-xl bg-transparent text-[rgb(var(--color-contrast-text))] transition-colors hover:bg-[rgb(var(--color-contrast-text)/0.15)]"
				onclick={() => moveMonth(1)}
				aria-label="Next month"
			>
				<span class="grid transition-transform duration-200 group-hover:translate-x-0.5 group-active:translate-x-1">
					<Icon icon={chevronRightIcon} width="20" height="20" />
				</span>
			</button>
		</div>
	</div>

	<div class="-mb-3 grid grid-cols-7 bg-[rgb(var(--color-primary))] px-1.5 pb-3 text-center text-sm font-medium text-[rgb(var(--color-contrast-text))]" aria-hidden="true">
		{#each ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'] as weekday}
			<span class="py-2">{weekday}</span>
		{/each}
	</div>
	<div class="calendar-dates overflow-hidden">
		{#key month}
			<div
				class:calendar-month-next={monthDirection > 0}
				class:calendar-month-previous={monthDirection < 0}
				class="calendar-month grid grid-cols-7 gap-0.5 sm:gap-1"
			>
				{#each Array(leadingDays) as _}
					<span></span>
				{/each}
				{#each Array(daysInMonth) as _, index}
					{@const day = index + 1}
					{@const date = keyFor(day)}
					<button
						type="button"
						disabled={!available.has(date)}
						onclick={() => (selected = date)}
						class={cellClass(date)}
						aria-label={date === today ? `${date} (today)` : date}
						aria-current={date === today ? 'date' : undefined}
						aria-pressed={selected === date}
					>
						{day}
						{#if date === today}
							<span class="pointer-events-none absolute bottom-1 left-1/2 size-1.5 -translate-x-1/2 rounded-full bg-current"></span>
						{/if}
					</button>
				{/each}
			</div>
		{/key}
	</div>
</div>

<style>
	.calendar-dates {
		border-radius: 0.75rem;
		padding: 0.75rem;
		background: color-mix(in srgb, rgb(var(--color-text)) 5%, rgb(var(--color-foreground)));
	}

	@media (min-width: 640px) {
		.calendar-dates {
			padding: 1rem;
		}
	}

	/* Hover pop for bookable days. :global because the class is applied via a
	   dynamic string in cellClass(), which Svelte's scoper can't see. */
	:global(.day-cell) {
		transition:
			transform 0.15s ease,
			background-color 0.15s ease,
			color 0.15s ease,
			box-shadow 0.15s ease;
	}
	:global(.day-cell:hover) {
		background: rgb(var(--color-primary) / 0.12) !important;
		color: rgb(var(--color-primary)) !important;
	}
	:global(.day-cell:active) {
		filter: brightness(0.92);
	}

	/* Arrow-shaped primary fill wipes in — same effect as the time slots. */
	:global(.day-cell)::before,
	:global(.calendar-selected)::before {
		content: '';
		position: absolute;
		inset: 0;
		right: 100%;
		background: rgb(var(--color-primary));
		clip-path: polygon(0 0, calc(100% - 8px) 0, 100% 50%, calc(100% - 8px) 100%, 0 100%);
		transition: right 0.25s ease;
		z-index: -1;
	}
	:global(.calendar-selected)::before {
		right: -8px;
	}
	:global(.calendar-selected)::after {
		content: '';
		position: absolute;
		inset: 0;
		border-radius: inherit;
		box-shadow: inset 0 0 0 4px rgb(var(--color-background) / 0.3);
		pointer-events: none;
	}

	.calendar-month {
		animation: calendar-month-next 260ms cubic-bezier(0.22, 1, 0.36, 1);
	}

	.calendar-month-previous {
		animation-name: calendar-month-previous;
	}

	.calendar-label {
		animation: calendar-label-next 260ms cubic-bezier(0.22, 1, 0.36, 1);
	}

	.calendar-label-previous {
		animation-name: calendar-label-previous;
	}

	@keyframes calendar-month-next {
		from {
			opacity: 0.5;
			transform: translateX(0.9rem);
		}
		to {
			opacity: 1;
			transform: translateX(0);
		}
	}

	@keyframes calendar-month-previous {
		from {
			opacity: 0.5;
			transform: translateX(-0.9rem);
		}
		to {
			opacity: 1;
			transform: translateX(0);
		}
	}

	@keyframes calendar-label-next {
		from {
			opacity: 0.5;
			transform: translateX(0.55rem);
		}
		to {
			opacity: 1;
			transform: translateX(0);
		}
	}

	@keyframes calendar-label-previous {
		from {
			opacity: 0.5;
			transform: translateX(-0.55rem);
		}
		to {
			opacity: 1;
			transform: translateX(0);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.calendar-month,
		.calendar-label {
			animation: none;
		}
	}
</style>
