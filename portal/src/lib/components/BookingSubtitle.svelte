<script lang="ts">
	let { text, id }: { text: string; id?: string } = $props();
</script>

<h2 {id} class="booking-subtitle">{text}</h2>

<style>
	.booking-subtitle {
		margin-top: 0.35rem;
		color: rgb(var(--color-text) / 0.65);
		font-size: 0.9375rem;
		font-weight: 400;
		line-height: 1.35;
	}
</style>
