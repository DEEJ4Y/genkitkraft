import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'GenKitKraft',
  tagline: 'Self-hostable LLM agent platform',
  favicon: 'img/favicon.ico',

  future: {
    v4: true,
  },

  url: 'https://DEEJ4Y.github.io',
  baseUrl: '/genkitkraft/',

  organizationName: 'DEEJ4Y',
  projectName: 'genkitkraft',
  trailingSlash: false,

  onBrokenLinks: 'throw',

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          editUrl:
            'https://github.com/DEEJ4Y/genkitkraft/tree/main/website/',
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    colorMode: {
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'GenKitKraft',
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'tutorialSidebar',
          position: 'left',
          label: 'Docs',
        },
        {
          href: 'https://github.com/DEEJ4Y/genkitkraft',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Documentation',
          items: [
            { label: 'Getting Started', to: '/docs/getting-started/installation' },
            { label: 'Guides', to: '/docs/guides/providers' },
            { label: 'Configuration', to: '/docs/configuration/environment-variables' },
            { label: 'API Reference', to: '/docs/api/overview' },
          ],
        },
        {
          title: 'More',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/DEEJ4Y/genkitkraft',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} GenKitKraft. Built with Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['bash', 'yaml', 'go', 'docker'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
