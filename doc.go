// Copyright: Jonathan Hall
// License: GNU AGPL, Version 3 or later; http://www.gnu.org/licenses/agpl.html

// Package anki provides a library to read *.apkg files produced by Anki
// (http://ankisrs.net/).
//
// A *.apkg file is simply a zip-compressed archive, which contains a the
// following files:
//
//	- collection.anki2      -- An SQLite3 database
//	- media                 -- A JSON blob mapping media file filenames to numbers
//	- [files numbered 0..n] -- Media files (referenced in the media file above)
//
// The SQLite3 Database contains the following tables. Detailed explanations
// of each column's use can be found inline below, in the struct definitions.
//
// 	CREATE TABLE col (
// 		id              integer primary key,
// 		crt             integer not null,
// 		mod             integer not null,
// 		scm             integer not null,
// 		ver             integer not null,
// 		dty             integer not null,
// 		usn             integer not null,
// 		ls              integer not null,
// 		conf            text not null,
// 		models          text not null,
// 		decks           text not null,
// 		dconf           text not null,
// 		tags            text not null
// 	);
//
// 	CREATE TABLE notes (
// 		id              integer primary key,   /* 0 */
// 		guid            text not null,         /* 1 */
// 		mid             integer not null,      /* 2 */
// 		mod             integer not null,      /* 3 */
// 		usn             integer not null,      /* 4 */
// 		tags            text not null,         /* 5 */
// 		flds            text not null,         /* 6 */
// 		sfld            integer not null,      /* 7 */
// 		csum            integer not null,      /* 8 */
// 		flags           integer not null,      /* 9 */
// 		data            text not null          /* 10 */
// 	);
//
// 	CREATE TABLE cards (
// 		id              integer primary key,   /* 0 */
// 		nid             integer not null,      /* 1 */
// 		did             integer not null,      /* 2 */
// 		ord             integer not null,      /* 3 */
// 		mod             integer not null,      /* 4 */
// 		usn             integer not null,      /* 5 */
// 		type            integer not null,      /* 6 */
// 		queue           integer not null,      /* 7 */
// 		due             integer not null,      /* 8 */
// 		ivl             integer not null,      /* 9 */
// 		factor          integer not null,      /* 10 */
// 		reps            integer not null,      /* 11 */
// 		lapses          integer not null,      /* 12 */
// 		left            integer not null,      /* 13 */
// 		odue            integer not null,      /* 14 */
// 		odid            integer not null,      /* 15 */
// 		flags           integer not null,      /* 16 */
// 		data            text not null          /* 17 */
// 	);
//
// 	CREATE TABLE graves (
// 		usn             integer not null,
// 		oid             integer not null,
// 		type            integer not null
// 	);
//
// 	CREATE TABLE revlog (
// 		id              integer primary key,
// 		cid             integer not null,
// 		usn             integer not null,
// 		ease            integer not null,
// 		ivl             integer not null,
// 		lastIvl         integer not null,
// 		factor          integer not null,
// 		time            integer not null,
// 		type            integer not null
// 	);
//
// When it is obvious that a column is no longer used by Anki, it is omitted
// from these Go data structures. When it is not obvious, it is included, but
// typically with a comment to the effect that its use is unknown. If you know
// of any inaccuracies or recent changes to the Anki schema, please create an
// issue.
package anki
