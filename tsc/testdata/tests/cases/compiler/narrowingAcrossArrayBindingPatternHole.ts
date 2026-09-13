// @strict: true

// The hole in the array binding pattern is a binding element without a name; the
// flow path from the narrowed use of `value` crosses its assignment node.
declare const pair: [number, string];

function narrowAcrossHole(value: string | number) {
    value = "text";
    const [, second] = pair;
    const narrowed: string = value;
    return narrowed + second;
}
