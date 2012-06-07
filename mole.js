#!/usr/bin/env node

// External modules.

var _ = require('underscore')._;
var fs = require('fs');
var kexec = require('kexec');
var temp = require('temp');
var util = require('util');

// Internal modules.

var sshConfig = require('./lib/ssh-config');
var expectConfig = require('./lib/expect-config');
var setupLocalIPs = require('./lib/setup-local-ips');

// Load a configuration, generate a temporary filename for ssh config.

var config = require('./specs.js').config;
config.sshConfig = temp.path({suffix: '.sshconfig'});

// Create and save the ssh config

var defaults = ['Host *', '  UserKnownHostsFile /dev/null', '  StrictHostKeyChecking no'].join('\n') + '\n';
var conf = defaults + sshConfig(config) + '\n';
fs.writeFileSync(config.sshConfig, conf);

// Create the expect script and save it to a temp file.

var expect = expectConfig(config) + '\n';
var expectFile = temp.path({suffix: '.expect'});
fs.writeFileSync(expectFile, expect);

// Set up local IP:s needed for forwarding and execute the expect scipt.

setupLocalIPs(config, function (c) {
    if (!c) {
        console.log('Failed to set up IP:s for forwarding.');
    } else {
        kexec('expect ' + expectFile);
    }
});

