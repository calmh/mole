"use strict";

var _ = require('underscore');
var colors = require('colors');

var isatty = process.stdout.isTTY;

// Figure out useful maximum column lengths, and create space padding we can
// use later.

var windowLen = process.stdout.getWindowSize()[0];
var maxColumnWidth = Math.floor(windowLen/2);
var space = '';
for (var i = 0; i < maxColumnWidth; i++) {
    space += ' ';
}

function out(str) {

    // Print the string, removing ANSI escape sequences if stdout is not a terminal.

    console.log(isatty ? str : str.stripColors);
}

module.exports = function (columns, data) {
    var maxLen = _.pluck(columns, 'length');
    var numCol = columns.length;
    var numRow = data.length;
    var totLen = 0;
    var row, r, c, val;
    var droppedColumn = false;

    // Calculate the maximum width of each column.

    for (r = 0; r < numRow; r++) {
        for (c = 0; c < numCol; c++) {
            maxLen[c] = Math.min(maxColumnWidth, Math.max(maxLen[c], data[r][c].stripColors.length));
        }
    }

    // Go through all columns and check that they fit in the current window.

    for (c = 0; c < numCol; c++) {

        // Increase the total length by the length of this column. If it's not the last one,
        // also allow for the two spacers that will be added.

        totLen += maxLen[c];
        if (c < numCol - 1) {
            totLen += 2;
        }

        // If we have exceeded the window length, decrease the number of
        // columns so this is the last one.  Also set the width of this columns
        // to fit in the remaining space.

        if (totLen > windowLen) {
            numCol = c + 1;
            maxLen[c] -= (totLen - windowLen)
        }

        // If the resulting maximum length was less than three characters, drop
        // the column entirely and set a flag indicating that. Normally we dont
        // need to indicate a dropped column since one or more of the truncated
        // values would have been marked.

        if (maxLen[c] < 3) {
            numCol--;
            droppedColumn = true;
        }
    }

    // Print the header row.

    row = '';
    for (c = 0; c < numCol; c++) {

        // We print headers in upper case.

        val = columns[c].toUpperCase();

        // If the length of the header exceeds the column width, truncate it
        // and add a bold red chevron.  Otherwise pad it with spaces to the
        // column width.

        if (val.length > maxLen[c]) {
            row += val.slice(0, maxLen[c] - 1).underline;
            row += '>'.red.bold;
        } else {
            row += val.underline;
            row += space.slice(0, maxLen[c] - columns[c].stripColors.length).underline;

            // If this is not the last column, add a spacer.

            if (c < numCol - 1) {
                row += '  ';
            }
        }
    }

    // If we dropped columns above without changing the column widths, add a
    // chevron here before printing the header row to indicate that.

    if (droppedColumn) {
        row += '>'.red.bold;
    }

    out(row);

    // Print all data rows.

    for (r = 0; r < numRow; r++) {
        row = '';

        for (c = 0; c < numCol; c++) {

            // If the data length exceeds the column width, truncate in the
            // same way as the headers above. Otherwise pad.

            if (data[r][c].stripColors.length > maxLen[c]) {
                val = data[r][c].stripColors.slice(0, maxLen[c] - 1) + '>'.red;
            } else {
                val = data[r][c] + space.slice(0, maxLen[c] - data[r][c].stripColors.length);
            }

            row += val;

            // If this is not the last column, add a spacer.

            if (c < numCol - 1) {
                row += '  ';
            }

        }

        out(row);
    }
};
