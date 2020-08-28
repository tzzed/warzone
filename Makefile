NAME := warzone

.PHONY: all $(NAME) insert-all-types insert-all-types-with-tx

all: $(NAME)

$(NAME):
	cd ./cmd/$@ && go install

insert-all-types:
	warzone -scenario insert-all-types -rm true > insert-all-types.csv

insert-all-types-with-tx:
	warzone -scenario insert-all-types-with-tx -rm true > insert-all-types-with-tx.csv
