// @flow
// Automatically generated by typewriter. Do not edit.
// http://www.github.com/natdm/typewriter


export type Embedded = { 
	name: string, 
	age: number
}

// Maps should all parse right.
// It's hard to get them to do that.
// @strict
export type Maps = {| 
	map_string_to_int: { [key: string]: number }, // I am a map of strings and ints

	map_string_to_ints: { [key: string]: Array<number> }, // I am a map of strings to a slice of ints

	map_string_to_maps: { [key: string]: { [key: string]: number } }// I am a map of strings to maps

|}

export type MyNumber = number

export type Names = Array<string>

export type Nested = { 
	person: Object
}

// OutgoingSocketMessage sends an action to the client that is easy
// for a redux store to parse
export type OutgoingSocketMessage = { 
	type: string, // action type

	payload: ?any, // action payload

	key: string// event key

}

export type People = { [key: string]: Person }

export type Person = { 
	name: string, 
	age: number
}

export type Something = { 
	some_map: { [key: string]: Array<Embedded> }
}

export type Thing = { 
	name: number
}
