import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  srcDir: '../',
  title: "Bab",
  description: "Zero dependencies. Cross-platform task runner for your projects.",
  head: [
    ['link', { rel: 'icon', href: '/favicon.ico' }]
  ],
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    nav: [
      { text: 'Home', link: '/' },
      { text: 'Get Started', link: '/get-started' },
      { text: 'Guide', link: '/syntax' },
      { text: 'GitHub', link: 'https://github.com/bab-sh/bab' }
    ],

    sidebar: [
      {
        text: 'Introduction',
        items: [
          { text: 'Getting Started', link: '/get-started' },
          { text: 'Installation', link: '/installation' }
        ]
      },
      {
        text: 'Guide',
        items: [
          { text: 'Babfile Syntax', link: '/syntax' },
          { text: 'Features', link: '/features' },
          { text: 'Script Compilation', link: '/compile' }
        ]
      }
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/bab-sh/bab' }
    ],

    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Made with ❤️ by aio for developers who value simplicity and reliability.'
    },

    editLink: {
      pattern: 'https://github.com/bab-sh/bab/edit/main/docs/vitepress/:path',
      text: 'Edit this page on GitHub'
    },

    search: {
      provider: 'local'
    }
  }
})
