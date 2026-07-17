import type { ThemeColors } from '$lib/types';

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
	const secondary = chroma * (1 - Math.abs((section % 2) - 1));
	const channels =
		section < 1 ? [chroma, secondary, 0] :
		section < 2 ? [secondary, chroma, 0] :
		section < 3 ? [0, chroma, secondary] :
		section < 4 ? [0, secondary, chroma] :
		section < 5 ? [secondary, 0, chroma] : [chroma, 0, secondary];
	const match = l - chroma / 2;
	return `#${channels.map((channel) => Math.round((channel + match) * 255).toString(16).padStart(2, '0')).join('')}`.toUpperCase();
}

function luminance(hex: string): number {
	const channels = Object.values(hexToRGB(hex)).map((channel) => {
		const value = channel / 255;
		return value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
	});
	return channels[0] * 0.2126 + channels[1] * 0.7152 + channels[2] * 0.0722;
}

function contrast(first: string, second: string): number {
	const values = [luminance(first), luminance(second)].sort((a, b) => b - a);
	return (values[0] + 0.05) / (values[1] + 0.05);
}

function accessibleText(backgrounds: string[], preferred?: string): string {
	if (preferred && backgrounds.every((background) => contrast(preferred, background) >= 4.5)) {
		return preferred;
	}
	const choices = ['#000000', '#FFFFFF'];
	return choices.sort((first, second) =>
		Math.min(...backgrounds.map((background) => contrast(second, background))) -
		Math.min(...backgrounds.map((background) => contrast(first, background)))
	)[0];
}

function entropy(range: number): number {
	return (Math.random() * 2 - 1) * range;
}

export function generateThemeColors(primary: string, mode: 'light' | 'dark'): ThemeColors {
	const { hue, saturation } = rgbToHSL(hexToRGB(primary));
	const shiftedHue = (offset: number) => (hue + offset + 360) % 360;
	const surfaceHue = (hue + 24 + entropy(5) + 360) % 360;
	const neutralSaturation = Math.min(18, Math.max(4, saturation * 0.18)) * (1 + entropy(0.08));
	const lightness = entropy(0.8);
	const foreground = mode === 'light'
		? hslToHex(shiftedHue(entropy(2)), neutralSaturation * 0.35, 99 + lightness * 0.35)
		: hslToHex(shiftedHue(entropy(2)), neutralSaturation, 17 + lightness);
	const background = mode === 'light'
		? hslToHex(surfaceHue, neutralSaturation, 96 + lightness)
		: hslToHex(surfaceHue, neutralSaturation, 10 + lightness);
	const preferredText = mode === 'light'
		? hslToHex(shiftedHue(entropy(2)), Math.min(22, saturation * 0.22), 17 + entropy(0.8))
		: hslToHex(shiftedHue(entropy(2)), Math.min(10, saturation * 0.1), 96 + entropy(0.8));

	return {
		primary: primary.toUpperCase(),
		primaryContrast: accessibleText([primary]),
		foreground,
		text: accessibleText([foreground, background], preferredText),
		background,
		border: mode === 'light'
			? hslToHex(surfaceHue, neutralSaturation, 55 + entropy(1.2))
			: hslToHex(surfaceHue, neutralSaturation, 52 + entropy(1.2)),
		shadow: mode === 'light'
			? hslToHex((hue + 180 + entropy(5)) % 360, neutralSaturation, 8 + entropy(0.8))
			: hslToHex((hue + 180 + entropy(5)) % 360, neutralSaturation, 2 + entropy(0.5))
	};
}
