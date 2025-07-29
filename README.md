# SilentRig

A secure, lightweight, self-hosted web platform for centralized Monero mining management and control.

## Overview

SilentRig is a self-hosted mining management platform similar to Cockpit or Netdata, designed specifically for Monero miners. It provides centralized control and monitoring of multiple mining rigs with a focus on security, performance, and ease of use.

## Features

### Core Functionality
- **Multi-Miner Management**: Control multiple XMRig, MoneroOcean, and other compatible miners through unified API/JSON-RPC interfaces
- **Intelligent Pool Switching**: Automatic switching between mining pools based on profitability and latency metrics
- **Real-time Monitoring**: Track temperature, power consumption, hashrates, and system health (for local deployments)

### User Interface
- **Dark Theme**: Easy-on-the-eyes interface optimized for 24/7 monitoring
- **Lightweight Design**: Fast, responsive interface that won't slow down your systems
- **Real-time Dashboard**: Live updates of all mining operations and statistics

## Technology Stack

- **Backend**: Go (Golang) - Fast, efficient, and secure
- **Frontend**: Vue.js/Svelte - Modern, reactive user interface
- **Database**: SQLite/Redis - Lightweight data storage and caching
- **APIs**: RESTful APIs with JSON-RPC support for miner communication

## Supported Miners

- **XMRig** - Full API support
- **MoneroOcean Miner** - Complete integration
- **SRBMiner-MULTI** - Basic support
- **XMR-Stak** - Legacy support

## Roadmap
- [ ] Advanced analytics and reporting
- [ ] Telegram/Discord notifications
- [ ] Custom mining strategies
- [ ] Hardware monitoring expansion
- [ ] Multi-algorithm support


## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Security

Security is a top priority. If you discover a security vulnerability, please open a public issue.

## Disclaimer

This software is for educational and legitimate mining purposes only. Users are responsible for complying with all applicable laws and regulations in their jurisdiction. The developers are not responsible for any misuse of this software.

---

**Built with ❤️ for the Monero mining community**
