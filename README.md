# Warzone

The Genji's warzone is a place where we monitor queries time complexity.

You can find in this [GDrive folder](https://drive.google.com/drive/u/0/folders/1bR4Sj1SEBmBxm3Mpl4bLzmoo-9ulFnW0?ths=true) some spreadsheets that contain charts. Each chart represents a scenario. A scenerio may contain one or many queries, it depends on what do we want to monitor. The X axis represents how many time the scenario has been ran and the Y axis represents the duration for each run in ms.

For now, we will generate a new spreadsheet every new release to see if it has a good or a bad impact.

## Usage

The warzone may be used locally by following the steps below.

First, clone the repository
```
git clone git@github.com:genjidb/warzone.git
```

Then, compile the program with the Makefile
```
// cd to the repository folder
make
```

The Makefile contains recipes for each scenario that will redirect the output to a csv file or you can directly use the `warzone` binary like this:
```
warzone -scenario insert-all-types -n 10000 -f 1000
```

For more options
```
warzone -h
```

## Contribution
Don't hesitate to suggest more scenarios!