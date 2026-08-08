import { chromium } from 'playwright'
const b = await chromium.launch()
const p = await b.newPage({ viewport: { width: 1440, height: 900 } })
await p.goto('http://localhost:6006/iframe.html?id=shell-platform--stacks-list&globals=theme:light&viewMode=story', { waitUntil: 'networkidle' })
await p.waitForTimeout(1500)
const box = (el) => el ? (({x,y,width,height,top,bottom,left,right}) => ({x,y,width,height,top,bottom,left,right}))(el.getBoundingClientRect()) : null
console.log(JSON.stringify(await p.evaluate(() => {
  const box = (el) => el ? (({x,y,width,height,top,bottom,left,right}) => ({x:+x.toFixed(1),y:+y.toFixed(1),w:+width.toFixed(1),h:+height.toFixed(1),top:+top.toFixed(1),bottom:+bottom.toFixed(1),left:+left.toFixed(1),right:+right.toFixed(1)}))(el.getBoundingClientRect()) : null
  const q = (s) => document.querySelector(s)
  const sidebar = q('[data-slot="sidebar"]') || q('[data-sidebar="sidebar"]')
  const inset = q('main') || q('[data-slot="sidebar-inset"]')
  const wordmark = [...document.querySelectorAll('span')].find(s => s.textContent.trim() === 'stackdome')
  const crumb = [...document.querySelectorAll('a,span')].find(s => s.textContent.trim() === 'Home')
  const trigger = q('[data-sidebar="trigger"]') || q('button')
  const firstNav = [...document.querySelectorAll('[data-sidebar="menu-button"], a')].find(a => a.textContent.trim() === 'Stacks')
  const groupLabel = [...document.querySelectorAll('div')].find(d => d.textContent.trim() === 'Platform' && d.children.length === 0)
  const sep = q('[data-slot="separator"]') || q('[data-orientation="horizontal"]')
  const cs = (el, ...props) => el ? Object.fromEntries(props.map(pr => [pr, getComputedStyle(el)[pr]])) : null
  return {
    sidebar: { ...box(sidebar), ...cs(sidebar, 'backgroundColor', 'borderRightWidth', 'borderRightColor') },
    inset: { ...box(inset), ...cs(inset, 'backgroundColor', 'borderRadius', 'borderColor', 'marginTop') },
    wordmark: box(wordmark),
    breadcrumbHome: box(crumb),
    trigger: box(trigger),
    navStacksRow: { ...box(firstNav), ...cs(firstNav, 'backgroundColor', 'borderRadius', 'fontSize') },
    groupLabelPlatform: { ...box(groupLabel), ...cs(groupLabel, 'fontSize', 'color', 'letterSpacing', 'textTransform') },
    body: cs(document.body, 'backgroundColor'),
  }
}), null, 2))
await b.close()
