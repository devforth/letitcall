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
		const base = 'relative aspect-square min-h-10 rounded-xl border-2 text-sm transition-colors';
		if (selected === date)
			return `${base} border-[rgb(var(--color-primary))] bg-[rgb(var(--color-primary))] font-semibold text-[rgb(var(--color-contrast-text))]`;
		if (date === today)
			// Today stands out with a bold primary number (plus the dot marker).
			return `${base} border-transparent font-bold text-[rgb(var(--color-primary))] ${available.has(date) ? 'hover:border-[rgb(var(--color-primary))] hover:bg-[rgb(var(--color-primary)/0.1)]' : 'cursor-not-allowed'}`;
		if (available.has(date))
			return `${base} border-[rgb(var(--color-border))] bg-[rgb(var(--color-foreground))] text-[rgb(var(--color-text))] hover:border-[rgb(var(--color-primary))] hover:bg-[rgb(var(--color-primary)/0.1)] hover:text-[rgb(var(--color-primary))]`;
		return `${base} border-transparent text-[rgb(var(--color-text)/0.35)]`;
	}
</script>

<div class="w-full" aria-label={monthLabel}>
	<div class="mb-5 grid grid-cols-[2.75rem_1fr_2.75rem] items-center">
		<button
			type="button"
			class="grid size-11 place-items-center rounded-xl border-2 border-[rgb(var(--color-border))] text-[rgb(var(--color-text))] transition-colors hover:border-[rgb(var(--color-primary))] hover:text-[rgb(var(--color-primary))] disabled:cursor-not-allowed disabled:opacity-30 disabled:hover:border-[rgb(var(--color-border))] disabled:hover:text-[rgb(var(--color-text))]"
			disabled={month <= minimumMonth}
			onclick={() => moveMonth(-1)}
			aria-label="Previous month"
		>
			<Icon icon={chevronLeftIcon} width="20" height="20" />
		</button>
		{#key month}
			<h2
				class="calendar-label text-center text-lg font-semibold text-[rgb(var(--color-text))]"
				class:calendar-label-next={monthDirection > 0}
				class:calendar-label-previous={monthDirection < 0}
			>
				{monthLabel}
			</h2>
		{/key}
		<button
			type="button"
			class="grid size-11 place-items-center rounded-xl border-2 border-[rgb(var(--color-border))] text-[rgb(var(--color-text))] transition-colors hover:border-[rgb(var(--color-primary))] hover:text-[rgb(var(--color-primary))]"
			onclick={() => moveMonth(1)}
			aria-label="Next month"
		>
			<Icon icon={chevronRightIcon} width="20" height="20" />
		</button>
	</div>

	{#key month}
		<div class:calendar-month-next={monthDirection > 0} class:calendar-month-previous={monthDirection < 0} class="calendar-month">
			<div class="grid grid-cols-7 text-center text-xs font-medium text-[rgb(var(--color-text)/0.6)]" aria-hidden="true">
				{#each ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'] as weekday}
					<span class="py-2">{weekday}</span>
				{/each}
			</div>
			<div class="grid grid-cols-7 gap-1">
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
		</div>
	{/key}
</div>

<style>
	.calendar-month {
		animation: calendar-month-next 160ms ease-out;
	}

	.calendar-month-previous {
		animation-name: calendar-month-previous;
	}

	.calendar-label {
		animation: calendar-label-next 160ms ease-out;
	}

	.calendar-label-previous {
		animation-name: calendar-label-previous;
	}

	@keyframes calendar-month-next {
		from {
			opacity: 0.8;
			transform: translateX(0.5rem);
		}
	}

	@keyframes calendar-month-previous {
		from {
			opacity: 0.8;
			transform: translateX(-0.5rem);
		}
	}

	@keyframes calendar-label-next {
		from {
			opacity: 0.8;
			transform: translateX(0.35rem);
		}
	}

	@keyframes calendar-label-previous {
		from {
			opacity: 0.8;
			transform: translateX(-0.35rem);
		}
	}
</style>
