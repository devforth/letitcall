import type { ThemeColors } from '$lib/types';
import { accessibleTextColor, contrastRatio, wcagAAContrast } from '$lib/color-contrast';

type RGB = { red: number; green: number; blue: number };

function hexToRGB(hex: string): RGB {
	return {
		red: Number.parseInt(hex.slice(1, 3), 16),
		green: Number.parseInt(hex.slice(3, 5), 16),
		blue: Number.parseInt(hex.slice(5, 7), 16)
	};
}

function rgbToHSL({ red, green, blue }: RGB) {
	const channels = [red / 255, green / 255, blue / 255];
	const maximum = Math.max(...channels);
	const minimum = Math.min(...channels);
	const lightness = (maximum + minimum) / 2;
	const delta = maximum - minimum;
	if (delta === 0) return { hue: 0, saturation: 0, lightness: lightness * 100 };

	const saturation = delta / (1 - Math.abs(2 * lightness - 1));
	let hue = 0;
	if (maximum === channels[0]) hue = 60 * (((channels[1] - channels[2]) / delta) % 6);
	if (maximum === channels[1]) hue = 60 * ((channels[2] - channels[0]) / delta + 2);
	if (maximum === channels[2]) hue = 60 * ((channels[0] - channels[1]) / delta + 4);
	return { hue: hue < 0 ? hue + 360 : hue, saturation: saturation * 100, lightness: lightness * 100 };
}

function hslToHex(hue: number, saturation: number, lightness: number): string {
	const s = saturation / 100;
	const l = lightness / 100;
	const chroma = (1 - Math.abs(2 * l - 1)) * s;
	const section = hue / 60;
	const middle = chroma * (1 - Math.abs((section % 2) - 1));
	const channels =
		section < 1 ? [chroma, middle, 0] :
		section < 2 ? [middle, chroma, 0] :
		section < 3 ? [0, chroma, middle] :
		section < 4 ? [0, middle, chroma] :
		section < 5 ? [middle, 0, chroma] : [chroma, 0, middle];
	const match = l - chroma / 2;
	return `#${channels.map((channel) => Math.round((channel + match) * 255).toString(16).padStart(2, '0')).join('')}`.toUpperCase();
}

function entropy(range: number): number {
	return (Math.random() * 2 - 1) * range;
}

function themedBackground(primary: string, mode: 'light' | 'dark', hue: number, saturation: number): string {
	const backgroundSaturation = Math.min(mode === 'light' ? 32 : 44, saturation * 0.55);
	const preferredLightness = mode === 'light' ? 94 + entropy(3) : 12 + entropy(3);
	const preferred = hslToHex(hue, backgroundSaturation, preferredLightness);
	if (contrastRatio(primary, preferred) >= wcagAAContrast) return preferred;

	const themedExtreme = mode === 'light' ? 100 : 0;
	const themedExtremeColor = hslToHex(hue, backgroundSaturation, themedExtreme);
	let passing = contrastRatio(primary, themedExtremeColor) >= wcagAAContrast
		? themedExtreme
		: mode === 'light' ? 0 : 100;
	let failing = preferredLightness;
	while (Math.abs(passing - failing) > 0.1) {
		const lightness = (passing + failing) / 2;
		const candidate = hslToHex(hue, backgroundSaturation, lightness);
		if (contrastRatio(primary, candidate) >= wcagAAContrast) passing = lightness;
		else failing = lightness;
	}
	return hslToHex(hue, backgroundSaturation, passing);
}

export function generateThemeColors(primary: string, mode: 'light' | 'dark'): ThemeColors {
	const { hue, saturation } = rgbToHSL(hexToRGB(primary));
	const shiftedHue = (offset: number) => (hue + offset + 360) % 360;
	const background = themedBackground(primary, mode, shiftedHue(entropy(12)), saturation);
	const preferredText = mode === 'light'
		? hslToHex(shiftedHue(entropy(2)), Math.min(22, saturation * 0.22), 17 + entropy(0.8))
		: hslToHex(shiftedHue(entropy(2)), Math.min(10, saturation * 0.1), 96 + entropy(0.8));

	return {
		primary: primary.toUpperCase(),
		text: accessibleTextColor([background], preferredText),
		background
	};
}
