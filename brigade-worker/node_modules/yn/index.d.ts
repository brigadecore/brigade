export interface Options {
	/**
	 * Use a key distance-based score to leniently accept typos of `yes` and `no`.
	 *
	 * @default false
	 */
	lenient?: boolean;

	/**
	 * Default value if no match was found.
	 *
	 * @default null
	 */
	default?: boolean | null;
}

/**
 * Parse yes/no like values.
 * The following case-insensitive values are recognized: `'y', 'yes', 'true', true, '1', 1, 'n', 'no', 'false', false, '0', 0`
 *
 * @param input - Value that should be converted.
 * @returns The parsed input if it can be parsed or the default value defined in the `default` option.
 */
export default function yn(input: any, options?: Options): boolean | null;
