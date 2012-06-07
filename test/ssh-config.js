var should = require('should');
var sshConfig = require('../lib/ssh-config');

describe('sshConfig', function () {
    it('should return a simple host config', function () {
        var config = {
            hosts: {
                test: {
                    addr: '1.2.3.4',
                    user: 'testuser',
                    port: 1234,
                    via: 'another-host',
                }
            }
        };

        var expected = [
            'Host test',
            '  User testuser',
            '  Hostname 1.2.3.4',
            '  Port 1234',
            '  ProxyCommand ssh another-host nc -w 1800 %h %p'
        ].join('\n');

        sshConfig(config).should.equal(expected);
    });

    it('should return multiple hosts', function () {
        var config = {
            hosts: {
                'another-host': {
                    addr: '2.2.3.4',
                    user: 'test1',
                },
                test: {
                    addr: '1.2.3.4',
                    user: 'testuser',
                    port: 1234,
                    via: 'another-host',
                }
            }
        };

        var expected = [
            'Host another-host',
            '  User test1',
            '  Hostname 2.2.3.4',
            'Host test',
            '  User testuser',
            '  Hostname 1.2.3.4',
            '  Port 1234',
            '  ProxyCommand ssh another-host nc -w 1800 %h %p'
        ].join('\n');

        sshConfig(config).should.equal(expected);
    });

    it('should return a config with forwards', function () {
        var config = {
            main: 'test',
            forwards: {
                'Foo': [
                    { from: '127.0.0.1:3994', to: '127.0.0.1:3994' },
                ],
                'Bar': [
                    { from: '127.0.0.1:42000', to: '10.0.0.5:42000' },
                    { from: '127.0.0.1:42002', to: '10.0.0.5:42002' },
                ]
            },
            hosts: {
                test: {
                    addr: '1.2.3.4',
                }
            }
        };

        var expected = [
            'Host test',
            '  Hostname 1.2.3.4',
            '  # Foo',
            '  LocalForward 127.0.0.1:3994 127.0.0.1:3994',
            '  # Bar',
            '  LocalForward 127.0.0.1:42000 10.0.0.5:42000',
            '  LocalForward 127.0.0.1:42002 10.0.0.5:42002',
        ].join('\n');

        sshConfig(config).should.equal(expected);
    });
});
