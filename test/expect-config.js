"use strict";

/*global it: false, describe: false */

var should = require('should');
var expectConfig = require('../lib/expect-config');

describe('expectConfig', function () {
    it('should spawn vpnc if there is a vpnc config file present', function () {
        var config = {
            sshConfig: '/tmp/sshconf',
            vpncConfig: '/tmp/vpncconf',
            main: 'test',
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

        expectConfig(config).should.match(/\bspawn sudo vpnc \/tmp\/vpncconf\n/);
    });
});
