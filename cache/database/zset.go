package database

import skiplist "github.com/sean-public/fast-skiplist"

type ZSetEntry struct {
	Val *skiplist.SkipList
}