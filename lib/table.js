var _ = require('underscore');

var maxColumnWidth = 40;
var space = '', line = '';
for (var i = 0; i < maxColumnWidth; i++) {
    space += ' ';
    line += '-';
}

module.exports = function (columns, data) {
    var maxLen = _.pluck(columns, 'length');
    var numCol = columns.length;
    var numRow = data.length;
    var row, r, c;

    for (r = 0; r < numRow; r++) {
        for (c = 0; c < numCol; c++) {
            maxLen[c] = Math.min(maxColumnWidth, Math.max(maxLen[c], data[r][c].length));
        }
    }

    row = '';
    for (c = 0; c < numCol; c++) {
        row += (columns[c] + space).slice(0, maxLen[c]).toUpperCase().underline;
        row += '  ';
    }
    console.log(row);

    for (r = 0; r < numRow; r++) {
        row = '';
        for (c = 0; c < numCol; c++) {
            if (data[r][c].length > maxLen[c]) {
                val = data[r][c].slice(0, maxLen[c] - 3) + '...';
            } else {
                val = (data[r][c] + space).slice(0, maxLen[c]);
            }

            if (c === 0) {
                val = val.green;
            }

            row += val + '  ';
        }
        console.log(row);
    }
}
