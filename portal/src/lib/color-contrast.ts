export const wcagAAContrast = 4.5;

function luminance(hex: string): number {
	const channels = [1, 3, 5].map((index) => {
		const value = Number.parseInt(hex.slice(index, index + 2), 16) / 255;
		return value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
	});
	return channels[0] * 0.2126 + channels[1] * 0.7152 + channels[2] * 0.0722;
}

export function contrastRatio(first: string, second: string): number {
	const values = [luminance(first), luminance(second)].sort((a, b) => b - a);
	return (values[0] + 0.05) / (values[1] + 0.05);
}

export function accessibleTextColor(backgrounds: string[], preferred?: string): string {
	if (preferred && backgrounds.every((background) => contrastRatio(preferred, background) >= wcagAAContrast)) {
		return preferred;
	}
	const choices = ['#000000', '#FFFFFF'];
	return choices.sort((first, second) =>
		Math.min(...backgrounds.map((background) => contrastRatio(second, background))) -
		Math.min(...backgrounds.map((background) => contrastRatio(first, background)))
	)[0];
}
