<script lang="ts">
	import Icon from '@iconify/svelte';
	import chevronLeftIcon from '@iconify-icons/tabler/chevron-left';
	import chevronRightIcon from '@iconify-icons/tabler/chevron-right';

	let {
		month = $bindable(),
		selected = $bindable(),
		availableDates,
		minimumMonth
	}: {
		month: string;
		selected: string;
		availableDates: string[];
		minimumMonth: string;
	} = $props();

	const [year, monthNumber] = $derived(month.split('-').map(Number));
	const daysInMonth = $derived(new Date(Date.UTC(year, monthNumber, 0)).getUTCDate());
	const leadingDays = $derived((new Date(Date.UTC(year, monthNumber - 1, 1)).getUTCDay() + 6) % 7);
	const available = $derived(new Set(availableDates));
	const monthLabel = $derived(
		new Intl.DateTimeFormat(undefined, { month: 'long', year: 'numeric', timeZone: 'UTC' }).format(
			new Date(Date.UTC(year, monthNumber - 1, 1))
		)
	);

	function moveMonth(amount: number) {
		const next = new Date(Date.UTC(year, monthNumber - 1 + amount, 1));
		month = `${next.getUTCFullYear()}-${String(next.getUTCMonth() + 1).padStart(2, '0')}`;
		selected = '';
	}

	function keyFor(day: number) {
		return `${month}-${String(day).padStart(2, '0')}`;
	}
</script>

<div class="w-full" aria-label={monthLabel}>
	<div class="mb-5 grid grid-cols-[2.75rem_1fr_2.75rem] items-center">
		<button
			type="button"
			class="grid size-11 place-items-center border border-black disabled:opacity-30"
			disabled={month <= minimumMonth}
			onclick={() => moveMonth(-1)}
			aria-label="Previous month"
		>
			<Icon icon={chevronLeftIcon} width="20" height="20" />
		</button>
		<h2 class="text-center text-lg font-semibold">{monthLabel}</h2>
		<button
			type="button"
			class="grid size-11 place-items-center border border-black"
			onclick={() => moveMonth(1)}
			aria-label="Next month"
		>
			<Icon icon={chevronRightIcon} width="20" height="20" />
		</button>
	</div>

	<div class="grid grid-cols-7 text-center text-xs font-medium" aria-hidden="true">
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
				class={`aspect-square min-h-10 border text-sm ${selected === date ? 'border-black bg-black text-white' : available.has(date) ? 'border-black bg-white text-black hover:bg-black hover:text-white' : 'border-transparent text-black/35'}`}
				aria-label={date}
				aria-pressed={selected === date}
			>
				{day}
			</button>
		{/each}
	</div>
</div>
