<script lang="ts">
	let { text, id }: { text: string; id?: string } = $props();
</script>

<h2 {id} class="booking-subtitle">{text}</h2>

<style>
	.booking-subtitle {
		margin-top: 0.375rem;
		font-size: 1.125rem;
		font-weight: 500;
		line-height: 1.4;
		opacity: 0.7;
	}
</style>
