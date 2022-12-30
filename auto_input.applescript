on run argv
    set theQuery to item 1 of argv
    set myArray to my theSplit(theQuery, "-")
    set theUsername to item 1 of myArray
    set thePassword to item 2 of myArray
    tell application "System Events"
	    keystroke theUsername
	    keystroke tab
	    keystroke thePassword
        -- keystroke return
    end tell
end run

on theSplit(theString, theDelimiter)
		-- save delimiters to restore old settings
		set oldDelimiters to AppleScript's text item delimiters
		-- set delimiters to delimiter to be used
		set AppleScript's text item delimiters to theDelimiter
		-- create the array
		set theArray to every text item of theString
		-- restore the old setting
		set AppleScript's text item delimiters to oldDelimiters
		-- return the result
		return theArray
end theSplit
