# SENG 468 Day Trading Project
## Group Members

## Notes
- always include cents (eg. 20.00)
- BUY -> QUOTE (60s to commit or cancel)
- COMMIT: only the last made buy will be committed
- CANCEL - only last made buy will be canceled (release buy)
- SELL -> QUOTE 
- COMMIT SELL, COMMIT BUY (same as buy)
- SET BUY AMOUNT - (several for specific stock) all of them will execute when trigger is activated
- CANCEL SET BUY - cancel set all of set buys and cancel trigger
- SET BUY TRIGGER - stock price to activate all SET BUY AMOUNTS
- one trigger per stock but multiple set buy/sell amounts possible
- each buy/sell amount added updates the trigger??


