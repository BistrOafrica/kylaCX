/**
 * Typed accessor for semantic design tokens.
 *
 * Components should usually use Tailwind utility classes (bg-surface,
 * text-fg, border-default, etc.). This module exists for the rare cases
 * where a token must be referenced from JS — chart libraries, inline
 * styles for dynamic values, or motion variants.
 *
 * The values returned are the CSS custom property *references* (e.g.
 * `var(--bg-surface)`), so they automatically follow the active theme.
 */

export const tokens = {
  bg: {
    canvas:   'var(--bg-canvas)',
    surface:  'var(--bg-surface)',
    elevated: 'var(--bg-elevated)',
    subtle:   'var(--bg-subtle)',
    muted:    'var(--bg-muted)',
    inverse:  'var(--bg-inverse)',
  },
  border: {
    default: 'var(--border-default)',
    strong:  'var(--border-strong)',
    subtle:  'var(--border-subtle)',
    focus:   'var(--border-focus)',
  },
  text: {
    primary:   'var(--text-primary)',
    secondary: 'var(--text-secondary)',
    muted:     'var(--text-muted)',
    disabled:  'var(--text-disabled)',
    inverse:   'var(--text-inverse)',
    link:      'var(--text-link)',
  },
  accent: {
    solid:  'var(--accent-solid)',
    hover:  'var(--accent-hover)',
    active: 'var(--accent-active)',
    subtle: 'var(--accent-subtle)',
    muted:  'var(--accent-muted)',
    fg:     'var(--accent-fg)',
  },
  status: {
    success: { solid: 'var(--status-success-solid)', subtle: 'var(--status-success-subtle)', fg: 'var(--status-success-fg)' },
    warn:    { solid: 'var(--status-warn-solid)',    subtle: 'var(--status-warn-subtle)',    fg: 'var(--status-warn-fg)' },
    danger:  { solid: 'var(--status-danger-solid)',  subtle: 'var(--status-danger-subtle)',  fg: 'var(--status-danger-fg)' },
    info:    { solid: 'var(--status-info-solid)',    subtle: 'var(--status-info-subtle)',    fg: 'var(--status-info-fg)' },
  },
  channel: {
    whatsapp:  'var(--channel-whatsapp)',
    email:     'var(--channel-email)',
    sms:       'var(--channel-sms)',
    voice:     'var(--channel-voice)',
    webchat:   'var(--channel-webchat)',
    instagram: 'var(--channel-instagram)',
    messenger: 'var(--channel-messenger)',
  },
  radius: {
    xs: 'var(--radius-xs)',
    sm: 'var(--radius-sm)',
    md: 'var(--radius-md)',
    lg: 'var(--radius-lg)',
    xl: 'var(--radius-xl)',
  },
  shadow: {
    e1: 'var(--shadow-1)',
    e2: 'var(--shadow-2)',
    e3: 'var(--shadow-3)',
    e4: 'var(--shadow-4)',
  },
  shell: {
    topbar:    'var(--shell-topbar)',
    statusbar: 'var(--shell-statusbar)',
    sidebar:   'var(--shell-sidebar)',
    aiRail:    'var(--shell-ai-rail)',
    list:      'var(--shell-list)',
  },
  font: {
    sans: 'var(--font-sans)',
    mono: 'var(--font-mono)',
  },
} as const;

export type Channel = keyof typeof tokens.channel;
export type StatusKind = keyof typeof tokens.status;
