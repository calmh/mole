"use strict";

exports.compress = function compressRange(list) {
    list.sort(function (a, b) { return a - b; });
    if (list.length < 3)
        return list;

    var newList = [];
    var start = list[0];
    var prev = list[0];
    for (var i = 1; i < list.length; i++) {
        if (list[i] !== prev + 1) {
            if (start === prev)
                newList.push(prev.toString());
            else
                newList.push(start + '-' + prev);
            start = list[i];
        }
        prev = list[i];
    }

    if (start === prev)
        newList.push(prev.toString());
    else
        newList.push(start + '-' + prev);

    return newList;
};

exports.expand = function expandRange(range) {
    if (range.indexOf('-') === -1)
        return [parseInt(range, 10)];
    var parts = range.split('-');
    var s = parseInt(parts[0], 10);
    var e = parseInt(parts[1], 10);
    var res = [];
    for (var i = s; i <= e; i++)
        res.push(i);
    return res;
};
