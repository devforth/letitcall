<script lang="ts">
	let { text }: { text: string } = $props();
</script>

<p class="booking-status-heading">{text}</p>

<style>
	.booking-status-heading {
		color: rgb(var(--color-text));
		font-size: clamp(1.625rem, 3vw, 2rem);
		font-weight: 600;
		letter-spacing: -0.025em;
		line-height: 1.1;
	}
</style>
