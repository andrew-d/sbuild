# Variables

- Global
	- Build dir: root directory that contains build products
	- Output dir: output directory for files
	- Cache dir: ${build_dir}/.cache - contains downloaded files

- Per-recipe:
	- Source dir: contains (possibly a copy of) the downloaded/fetched sources

# Building

- Given a recipe
	- Recursively find all dependencies and topologically sort them
	- Run individual builds in order

## Build Steps (for a single recipe)

- Remove and re-create the source directory
- For each source:
	- If it's not in the cache, download it there
	- Verify the checksum
		- Remove the source from the cache if it fails, so it can get re-downloaded
			next time.
	- Copy (or link?) the source from the cache into the recipe source dir
- Create the per-recipe environment
	- TODO
- Run the steps the following order:
	- Prepare
	- Build
	- Finalize

## Environments

- We have a basic environment, which is from the OS
- We add to this by overriding the following variables
		AR           := ${CROSS_PREFIX}-ar
		CC           := ${CROSS_PREFIX}-gcc
		CXX          := ${CROSS_PREFIX}-g++
		LD           := ${CROSS_PREFIX}-ld
		RANLIB       := ${CROSS_PREFIX}-ranlib
		STRIP        := ${CROSS_PREFIX}-strip
- Finally, we need to insert a per-recipe environment, containing flags from
	all the recipe's dependencies.
	- A recipe can specify flags that are to be inserted into the environment of
		its dependants.
