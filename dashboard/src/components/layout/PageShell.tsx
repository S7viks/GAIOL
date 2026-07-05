import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'

type PageHeaderProps = {
  title: string
  description?: ReactNode
  actions?: ReactNode
}

export function PageHeader({ title, description, actions }: PageHeaderProps) {
  return (
    <header className="page-header">
      <div className="page-header__main">
        <h1>{title}</h1>
        {description ? <p className="page-header__desc">{description}</p> : null}
      </div>
      {actions ? <div className="page-header__actions">{actions}</div> : null}
    </header>
  )
}

type PageSectionProps = {
  title?: string
  subtitle?: ReactNode
  children: ReactNode
  className?: string
  headerActions?: ReactNode
}

export function PageSection({ title, subtitle, children, className = '', headerActions }: PageSectionProps) {
  return (
    <section className={`panel page-section ${className}`.trim()}>
      {(title || headerActions) && (
        <div className="page-section__head">
          <div>
            {title ? <h2 className="page-section__title">{title}</h2> : null}
            {subtitle ? <p className="page-section__subtitle">{subtitle}</p> : null}
          </div>
          {headerActions ? <div className="page-section__actions">{headerActions}</div> : null}
        </div>
      )}
      {children}
    </section>
  )
}

type PageAlertProps = {
  variant?: 'warn' | 'err'
  title?: string
  children: ReactNode
  actionTo?: string
  actionLabel?: string
}

export function PageAlert({ variant = 'warn', title, children, actionTo, actionLabel }: PageAlertProps) {
  return (
    <div className={`alert alert--${variant} page-alert`}>
      <div className="page-alert__body">
        {title ? <strong>{title}</strong> : null}
        <div className="page-alert__content">{children}</div>
      </div>
      {actionTo && actionLabel ? (
        <Link to={actionTo} className="btn btn--secondary btn--sm">
          {actionLabel}
        </Link>
      ) : null}
    </div>
  )
}

export function PageStack({ children, className = '' }: { children: ReactNode; className?: string }) {
  return <div className={`page-stack ${className}`.trim()}>{children}</div>
}
