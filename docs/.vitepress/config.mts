import { defineConfig } from 'vitepress'

const ogUrl = 'https://docs.bab.sh'
const ogImage = 'https://cdn.bab.sh/l/logo'
const ogTitle = 'Bab - Custom commands for every project'
const ogDescription = 'A modern task runner built for seamless development workflows. Zero dependencies, cross-platform, simple and powerful.'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  srcDir: './content',
  title: "Bab",
  description: ogDescription,

  head: [
    ['link', { rel: 'icon', type: 'image/png', href: 'https://cdn.bab.sh/l/favicon' }],
    ['link', { rel: 'apple-touch-icon', href: 'https://cdn.bab.sh/l/icon' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:url', content: ogUrl }],
    ['meta', { property: 'og:title', content: ogTitle }],
    ['meta', { property: 'og:description', content: ogDescription }],
    ['meta', { property: 'og:image', content: ogImage }],
    ['meta', { name: 'twitter:card', content: 'summary_large_image' }],
    ['meta', { name: 'twitter:title', content: ogTitle }],
    ['meta', { name: 'twitter:description', content: ogDescription }],
    ['meta', { name: 'twitter:image', content: ogImage }],
    ['meta', { name: 'theme-color', content: '#646cff' }],
  ],

  themeConfig: {
    logo: 'https://cdn.bab.sh/l/icon',

    nav: [
      { text: 'Home', link: '/' },
      { text: 'Guide', link: '/guide/getting-started' },
      {
        text: 'Reference',
        items: [
          { text: 'CLI Reference', link: '/guide/cli-reference' },
          { text: 'Roadmap', link: '/reference/roadmap' }
        ]
      },
      { text: 'GitHub', link: 'https://github.com/bab-sh/bab' }
    ],

    sidebar: [
      {
        text: 'Introduction',
        items: [
          { text: 'Getting Started', link: '/guide/getting-started' },
          { text: 'Installation', link: '/guide/installation' },
          { text: 'Updating', link: '/guide/updating' }
        ]
      },
      {
        text: 'Guide',
        items: [
          { text: 'Babfile Syntax', link: '/guide/babfile-syntax' },
          { text: 'CLI Reference', link: '/guide/cli-reference' }
        ]
      },
      {
        text: 'Reference',
        items: [
          { text: 'Roadmap', link: '/reference/roadmap' },
          { text: 'Contributing', link: '/contributing' }
        ]
      }
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/bab-sh/bab' },
      { icon: 'discord', link: 'https://discord.bab.sh' },
      { icon: 'x', link: 'https://x.com/babshdev' },
      { icon: 'instagram', link: 'https://instagram.com/babshdev' },
      { icon: 'reddit', link: 'https://reddit.com/r/babsh' },
      { icon: 'threads', link: 'https://threads.net/@babshdev' }
    ],

    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Built with ❤️ by AIO for developers who value simplicity and reliability.'
    },

    editLink: {
      pattern: 'https://github.com/bab-sh/bab/edit/main/docs/content/:path',
      text: 'Edit this page on GitHub'
    },

    search: {
      provider: 'local'
    },

    outline: {
      level: [2, 3],
      label: 'On this page'
    }
  }
})
