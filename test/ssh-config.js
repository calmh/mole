"use strict";

/*global it: false, describe: false */

var should = require('should');
var sshConfig = require('../lib/ssh-config');

describe('sshConfig', function () {
    it('should return a simple host config', function () {
        var config = {
            general: { },
            sshConfig: '/tmp/sshconf',
            hosts: {
                test: {
                    addr: '1.2.3.4',
                    user: 'testuser',
                    port: 1234,
                    password: 'something',
                    via: 'another-host'
                }
            }
        };

        var expected = [
            /Host test/,
            /User testuser/,
            /Hostname 1.2.3.4/,
            /Port 1234/,
            /ProxyCommand ssh -F \/tmp\/sshconf another-host nc -w 1800 %h %p/,
            /PubkeyAuthentication no/,
            /PasswordAuthentication yes/
        ];

        expected.forEach(function (line) {
            sshConfig(config).should.match(line);
        });
    });

    it('should return a host config with key', function () {
        var config = {
            general: { },
            sshConfig: '/tmp/sshconf',
            hosts: {
                test: {
                    addr: '1.2.3.4',
                    user: 'testuser',
                    port: 1234,
                    key: 'something'
                }
            }
        };

        var expected = [
            /Host test/,
            /User testuser/,
            /Hostname 1.2.3.4/,
            /Port 1234/,
            /PubkeyAuthentication yes/,
            /PasswordAuthentication no/,
            /IdentityFile \//
        ];

        expected.forEach(function (line) {
            sshConfig(config).should.match(line);
        });
    });

    it('should return multiple hosts', function () {
        var config = {
            general: { },
            sshConfig: '/tmp/sshconf',
            hosts: {
                'another-host': {
                    addr: '2.2.3.4',
                    user: 'test1',
                    password: 'something'
                },
                test: {
                    addr: '1.2.3.4',
                    user: 'testuser',
                    port: 1234,
                    via: 'another-host',
                    password: 'something'
                }
            }
        };

        var expected = [
            /Host another-host/,
            /User test1/,
            /Hostname 2.2.3.4/,
            /Host test/,
            /User testuser/,
            /Hostname 1.2.3.4/,
            /Port 1234/,
            /ProxyCommand ssh -F \/tmp\/sshconf another-host nc -w 1800 %h %p/
        ];

        expected.forEach(function (line) {
            sshConfig(config).should.match(line);
        });
    });

    it('should return a config with forwards', function () {
        var config = {
            general: {
                main: 'test'
            },
            forwards: {
                'Foo': { '127.0.0.1:3994': '127.0.0.1:3994' },
                'Bar': {
                    '127.0.0.1:42000': '10.0.0.5:42000',
                    '127.0.0.1:42002': '10.0.0.5:42002'
                }
            },
            hosts: {
                test: {
                    addr: '1.2.3.4',
                    user: 'something',
                    password: 'something'
                }
            }
        };

        var expected = [
            /Host test/,
            /User something/,
            /Hostname 1.2.3.4/,
            /# Foo/,
            /LocalForward 127.0.0.1:3994 127.0.0.1:3994/,
            /# Bar/,
            /LocalForward 127.0.0.1:42000 10.0.0.5:42000/,
            /LocalForward 127.0.0.1:42002 10.0.0.5:42002/
        ];

        expected.forEach(function (line) {
            sshConfig(config).should.match(line);
        });
    });
});
