"use strict";

var _ = require('underscore');
var fs = require('fs');

module.exports = Users;

function Users(fname) {
    this.store = fname;
    try {
        this.users = JSON.parse(fs.readFileSync(fname));
    } catch (err) {
        this.users = {};
    }
}

Users.prototype.set = function (user, obj) {
    this.users[user] = obj;
    this.save();
};

Users.prototype.get = function (user) {
    return this.users[user];
};

Users.prototype.del = function (user) {
    delete this.users[user];
};

Users.prototype.saveNow = function (callback) {
    if (!this || !this.users || !this.store) {
        // Scheduled on an object that doesn't exist.
        return;
    }
    if (this.saveTimeout) {
        clearTimeout(this.saveTimeout);
        delete this.saveTimeout;
    }
    fs.writeFile(this.store, JSON.stringify(this.users, null, '  '), callback);
};

Users.prototype.save = function () {
    if (this.saveTimeout) {
        clearTimeout(this.saveTimeout);
        delete this.saveTimeout;
    }
    this.saveTimeout = setTimeout(Users.prototype.saveNow.bind(this), 200);
};

Users.prototype.all = function () {
    return _.map(this.users, function (data, name) {
        return { name: name, data: data }
    });
};
