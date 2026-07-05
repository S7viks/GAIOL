import { Link, Outlet } from 'react-router-dom'

export function PublicLayout() {
  return (
    <>
      <nav className="public-nav" aria-label="Main navigation">
        <div className="public-nav__inner">
          <Link to="/" className="public-nav__brand">
            gaiol<span className="public-nav__brand-accent">_</span>
          </Link>
          <div className="public-nav__links">
            <a href="/#dispatch">dispatch</a>
            <a href="/#compare">diff</a>
            <Link to="/login">login</Link>
            <Link to="/home">dashboard</Link>
          </div>
          <Link to="/signup" className="public-nav__cta">
            Get your key
          </Link>
        </div>
      </nav>
      <Outlet />
    </>
  )
}
