#!/usr/bin/env node

// We now use strict for everything since Node supports it.

"use strict";

// Require a whole bunch of external libraries.

var _ = require('underscore');
var colors = require('colors');
var fs = require('fs');
var inireader = require('inireader');
var mkdirp = require('mkdirp');
var parser = require('nomnom');
var path = require('path');

// Figure out if we are running with a TTY on stdout or not.
// If not, we'll avoid using ANSI colors later on.

var isatty = process.stdout.isTTY;

// The existsSync function moved between Node 0.6 and 0.8. We monkeypatch fs if
// it's not already there.
//
// You'll be seeing a lot of `*Sync` calls in here. If that disturbs you, keep
// in mind that this is a CLI program that runs once and then exits, not some
// sort of high performance IO-bound server code, mmkay?

if (!fs.existsSync) {
    fs.existsSync = path.existsSync;
}

var state = {};

// We load our own package file to get at the version number.

state.pkg = require(path.join(__dirname, 'package.json'));

// Load internal modules.

var con = require('./lib/console');
var init = require('./lib/init');
var Client = require('./lib/client');

// Prevent running as root.
if (process.getuid && process.getuid() === 0) {
    con.fatal('Do not run mole as root');
}

state.client = new Client();

// All server errors are fatal.

state.client.on('error', function (err) {
    con.fatal(err);
});

// Make a best guess of the user's home directory, with fallback to
// /tmp/<username>.

var homeDir = path.resolve(process.env.HOME || process.env.USERPROFILE ||
                           path.join('/tmp', process.env.USER || process.env.LOGNAME));

// Set up variables pointing to our config directory, certificate files and
// subdirectories for tunnels and packages.

state.path = { configDir: path.join(homeDir, '.mole') };
state.path.certFile = path.join(state.path.configDir, 'mole.crt');
state.path.configFile = path.join(state.path.configDir, 'mole.ini');
state.path.keyFile = path.join(state.path.configDir, 'mole.key');
state.path.pkgDir = path.join(state.path.configDir, 'pkg');
state.path.tunnels = path.join(state.path.configDir, 'tunnels');

// Create the tunnel and package directories. Any needed components leading up
// to these directories will be created as well as needed. No harm if they
// already exist.

mkdirp.sync(state.path.tunnels);
mkdirp.sync(state.path.pkgDir);

// Mark the entire config directory as private since we'll be storing keys and
// passwords in plaintext in there.

fs.chmodSync(state.path.configDir, 448 /* 0700 octal */);

// Load the config file. If it doesn't exist, set defaults and write a new
// config file.

state.config = new inireader.IniReader();
try {
    state.config.load(state.path.configFile);
} catch (err) {
    con.info('No config, using defaults.');
    state.config.param('server.port', 9443);
    state.config.write();
}

// Set up the help text that will be appended after the commands and options
// summary when the user makes an error or runs mole without parameters.

var helptext = [
    'Version:',
    '  mole v' + state.pkg.version + '\t(https://github.com/calmh/mole)',
    '  node ' + process.version,
    '',
    'Examples:',
    '',
    'Register with server "mole.example.com" and a token:',
    '  mole register mole.example.com 80721953-b4f2-450e-aaf4-a1c0c7599ec2'.bold,
    '',
    'List available tunnels:',
    '  mole list'.bold,
    '',
    'Dig a tunnel to "operator3":',
    '  mole dig operator3'.bold,
    '',
    'Fetch new and updated tunnel specifications from the server:',
    '  mole pull'.bold
].join('\n');

// Set the name of our 'script'.

parser.script('mole');

// Load all command modules. Add them into an array with prio and name as the
// first elements to allow easy sorting. Sort by prio, name.

var cmds = fs.readdirSync(path.join(__dirname, 'cmd'))
.map(function (module) {
    if (!module.match(/^[a-z0-9]+\.js$/)) { return null; }

    var cmd = require('./cmd/' + module);
    var name = path.basename(module, '.js');
    return [ cmd.prio || 5, name, cmd ];
})
.sort();

// Add them to the command line parser.

cmds.forEach(function (arr) {
    if (!arr) { return; }

    var name = arr[1];
    var module = arr[2];
    var cmdp = parser.command(name);

    cmdp.help(module.help);

    _.each(module.options, function (v, k) {
        cmdp.option(k, v);
    });

    cmdp.callback(function (opts) {
        init(opts, state);
        module(opts, state);
    });
});

// `-d` always turns on debug.

parser.option('debug', { abbr: 'd', flag: true, help: 'Display debug output' });

// `-h` shows help. This is actually implemented totally by `nomnom`, but we
// need to define the option so it shows up in the usage information.

parser.option('help', { abbr: 'h', flag: true, help: 'Display command help' });

// Add the help text, with or without colors depending on the TTY status.

parser.help(isatty ? helptext : helptext.stripColors);

// Parse command line arguments. This will call the defined callbacks for matching commands.

parser.parse();

