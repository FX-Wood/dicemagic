module github.com/fx-wood/dicemagic/internal/dicelang/cli

go 1.12

require (
	github.com/fx-wood/dicemagic/internal/dicelang v0.1.0
	github.com/fx-wood/dicemagic/internal/dicelang/errors v0.1.0
)

replace github.com/fx-wood/dicemagic/internal/dicelang v0.1.0 => ../

replace github.com/fx-wood/dicemagic/internal/dicelang/errors v0.1.0 => ../errors
