#!/usr/bin/env node

var cli = require('cli').enable('status');

cli.parse({ sshdebug: [ 'v', 'Enable verbose ssh sessions' ] }, ['connect', 'list']);

cli.main(function(args, options) {
    if (cli.command === 'connect' && !this.argc) {
        cli.fatal('Usage: ' + cli.app + ' connect <destination>');
    }

    // External modules.

    cli.debug('Starting up');
    var _ = require('underscore')._;
    var fs = require('fs');
    var kexec = require('kexec');
    var temp = require('temp');
    var util = require('util');

    // Internal modules.

    cli.debug('Loading modules');
    var sshConfig = require('./lib/ssh-config');
    var expectConfig = require('./lib/expect-config');
    var setupLocalIPs = require('./lib/setup-local-ips');

    // Load a configuration, generate a temporary filename for ssh config.

    cli.debug('Loading configuration');
    var config = require('./' + args[0]).config;
    config.sshConfig = temp.path({suffix: '.sshconfig'});

    // Create and save the ssh config

    cli.debug('Creating SSH configuration');
    var defaults = ['Host *', '  UserKnownHostsFile /dev/null', '  StrictHostKeyChecking no'].join('\n') + '\n';
    var conf = defaults + sshConfig(config) + '\n';
    fs.writeFileSync(config.sshConfig, conf);
    cli.debug('SSH configuration was saved to ' + config.sshConfig);
    cli.ok('Created SSH configuration');

    // Set up local IP:s needed for forwarding and execute the expect scipt.

    cli.debug('Setting up local IP:s for forwarding');
    setupLocalIPs(config, cli, options, function (c) {
        if (!c) {
            cli.error('Failed to set up IP:s for forwarding');
            cli.info('Continuing without forwarding');
            delete config.forwards;
        }

        // Create the expect script and save it to a temp file.

        cli.debug('Creating Expect script');
        var expect = expectConfig(config, cli, options) + '\n';
        var expectFile = temp.path({suffix: '.expect'});
        fs.writeFileSync(expectFile, expect);
        cli.debug('Expect script was saved to ' + expectFile);
        cli.ok('Created Expect script');

        cli.ok('Shifting into cyberspace');
        kexec('expect ' + expectFile);
    });
});

