"use strict";

var _ = require('underscore');
var fs = require('fs');

module.exports = JSONStore;

function JSONStore(fname) {
    this.store = fname;
    try {
        this.items = JSON.parse(fs.readFileSync(fname));
    } catch (err) {
        this.items = {};
    }
}

JSONStore.prototype.set = function (key, obj) {
    this.items[key] = obj;
    this.save();
};

JSONStore.prototype.get = function (key) {
    return this.items[key];
};

JSONStore.prototype.del = function (key) {
    delete this.items[key];
};

JSONStore.prototype.saveNow = function (callback) {
    if (!this || !this.items || !this.store) {
        // Scheduled on an object that doesn't exist.
        return;
    }
    if (this.saveTimeout) {
        clearTimeout(this.saveTimeout);
        delete this.saveTimeout;
    }
    fs.writeFile(this.store, JSON.stringify(this.items, null, '  '), callback);
};

JSONStore.prototype.save = function () {
    if (this.saveTimeout) {
        clearTimeout(this.saveTimeout);
        delete this.saveTimeout;
    }
    this.saveTimeout = setTimeout(JSONStore.prototype.saveNow.bind(this), 200);
};

JSONStore.prototype.all = function () {
    return _.map(this.items, function (data, name) {
        return { name: name, data: data }
    });
};
