import { defineConfig } from 'vitepress'

// pptx-go documentation site. Build-time only (Node/VitePress); never part of
// the Go module (P4). Deployed to GitHub Pages by .github/workflows/pages.yml.
export default defineConfig({
  title: 'pptx-go',
  description: 'A pure-Go library for authoring and reading PowerPoint (PPTX) files.',
  lang: 'en-US',
  // GitHub Pages serves a project site under /<repo>/.
  base: '/pptx-go/',
  cleanUrls: true,
  lastUpdated: true,
  themeConfig: {
    nav: [
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'Scene catalog', link: '/catalog/' },
      { text: 'API reference', link: '/reference/pptx' },
      { text: 'GitHub', link: 'https://github.com/hurtener/pptx-go' },
    ],
    sidebar: {
      '/guide/': [
        {
          text: 'Guide',
          items: [
            { text: 'Getting started', link: '/guide/getting-started' },
            { text: 'The builder (pptx)', link: '/guide/builder' },
            { text: 'Themes & tokens', link: '/guide/theme' },
            { text: 'The scene renderer', link: '/guide/scene' },
            { text: 'Assets, icons & rasters', link: '/guide/assets' },
            { text: 'Reading decks', link: '/guide/reading' },
          ],
        },
      ],
      '/catalog/': [
        {
          text: 'Scene catalog',
          items: [
            { text: 'Overview', link: '/catalog/' },
            { text: 'Text leaves', link: '/catalog/text-leaves' },
            { text: 'Visual leaves', link: '/catalog/visual-leaves' },
            { text: 'Asset leaves', link: '/catalog/asset-leaves' },
            { text: 'Containers', link: '/catalog/containers' },
          ],
        },
      ],
      '/reference/': [
        {
          text: 'API reference',
          items: [
            { text: 'pptx (builder)', link: '/reference/pptx' },
            { text: 'scene (renderer)', link: '/reference/scene' },
            { text: 'Design decisions', link: '/reference/decisions' },
          ],
        },
      ],
    },
    socialLinks: [
      { icon: 'github', link: 'https://github.com/hurtener/pptx-go' },
    ],
    footer: {
      message: 'Apache-2.0 licensed.',
      copyright: 'pptx-go — github.com/hurtener/pptx-go',
    },
    search: { provider: 'local' },
  },
})
