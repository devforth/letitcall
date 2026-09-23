type RGB = {
	red: number;
	green: number;
	blue: number;
};

function initialsHash(initials: string): number {
	let hash = 0;
	for (const character of initials) {
		hash = (hash * 31 + character.codePointAt(0)!) >>> 0;
	}
	return hash;
}

function hslToRGB(hue: number, saturation: number, lightness: number): RGB {
	const chroma = (1 - Math.abs(2 * lightness - 1)) * saturation;
	const section = hue / 60;
	const middle = chroma * (1 - Math.abs((section % 2) - 1));
	const channels =
		section < 1 ? [chroma, middle, 0] :
		section < 2 ? [middle, chroma, 0] :
		section < 3 ? [0, chroma, middle] :
		section < 4 ? [0, middle, chroma] :
		section < 5 ? [middle, 0, chroma] : [chroma, 0, middle];
	const match = lightness - chroma / 2;
	return {
		red: Math.round((channels[0] + match) * 255),
		green: Math.round((channels[1] + match) * 255),
		blue: Math.round((channels[2] + match) * 255)
	};
}

function hexToRGB(hex: string): RGB {
	return {
		red: Number.parseInt(hex.slice(1, 3), 16),
		green: Number.parseInt(hex.slice(3, 5), 16),
		blue: Number.parseInt(hex.slice(5, 7), 16)
	};
}

function luminance({ red, green, blue }: RGB): number {
	return [red, green, blue]
		.map((channel) => {
			const value = channel / 255;
			return value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
		})
		.reduce((sum, value, index) => sum + value * [0.2126, 0.7152, 0.0722][index], 0);
}

function contrast(first: RGB, second: RGB): number {
	const [light, dark] = [luminance(first), luminance(second)].sort((a, b) => b - a);
	return (light + 0.05) / (dark + 0.05);
}

export function avatarColorsFromInitials(initials: string, text: string, background: string) {
	const hash = initialsHash(initials);
	const avatarBackground = hslToRGB(hash % 360, 0.55 + ((hash >>> 9) % 16) / 100, 0.42 + ((hash >>> 17) % 12) / 100);
	const textColor = contrast(avatarBackground, hexToRGB(text)) >= contrast(avatarBackground, hexToRGB(background))
		? 'rgb(var(--color-text))'
		: 'rgb(var(--color-background))';

	return {
		background: `rgb(${avatarBackground.red} ${avatarBackground.green} ${avatarBackground.blue})`,
		textColor
	};
}
