import { useEffect } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import type { AppDispatch, RootState } from './store';
import { fetchRecipes } from '../contexts/recipes';

export function App() {
  const dispatch = useDispatch<AppDispatch>();
  const { items, status, error } = useSelector((state: RootState) => state.recipes);

  useEffect(() => {
    if (status === 'idle') {
      void dispatch(fetchRecipes());
    }
  }, [dispatch, status]);

  return (
    <main className="page-shell">
      <header className="hero">
        <div className="hero-copy">
          <p className="eyebrow">KITCHENHELPER · DEINE SAMMLUNG</p>
          <h1>Was kochen wir heute?</h1>
          <p className="hero-description">
            Bewährte Lieblingsgerichte und neue Ideen, an einem Ort gesammelt.
          </p>
        </div>
        <div className="recipe-count" aria-label={`${items.length} Rezepte`}>
          <span>{items.length}</span>
          <small>Rezepte</small>
        </div>
      </header>

      <section className="recipe-section" aria-labelledby="recipes-heading">
        <div className="section-heading">
          <div>
            <p className="eyebrow">GUT &amp; LECKER</p>
            <h2 id="recipes-heading">Alle Rezepte</h2>
          </div>
          <span className="recipe-total">{items.length} insgesamt</span>
        </div>

        {status === 'loading' && <p className="notice">Rezepte werden geladen …</p>}
        {status === 'failed' && (
          <div className="notice error" role="alert">
            {error}
          </div>
        )}
        {status === 'succeeded' && items.length === 0 && (
          <div className="empty-state">
            <span className="empty-icon" aria-hidden="true">✦</span>
            <h3>Noch ist hier Platz am Tisch.</h3>
            <p>Deine Rezeptsammlung ist noch leer. Schau bald wieder vorbei.</p>
          </div>
        )}
        {items.length > 0 && (
          <div className="recipe-grid">
            {items.map((recipe) => (
              <article className="recipe-card" key={recipe.id}>
                <div className="card-art" aria-hidden="true">
                  <span>{recipe.name.slice(0, 1).toLocaleUpperCase('de-DE')}</span>
                </div>
                <div className="card-content">
                  <div className="card-meta">
                    <span>REZEPT</span>
                    <span>v{recipe.version}</span>
                  </div>
                  <h3>{recipe.name}</h3>
                  <p className="card-description">
                    {recipe.description || recipe.text || 'Ein Rezept für jeden Anlass.'}
                  </p>
                  <div className="card-footer">
                    <span>{recipe.author || 'KitchenHelper'}</span>
                    <span>{recipe.ingredients.length} Zutaten</span>
                  </div>
                  {recipe.tags.length > 0 && (
                    <ul className="tags" aria-label="Schlagwörter">
                      {recipe.tags.slice(0, 3).map((tag) => (
                        <li key={tag}>{tag}</li>
                      ))}
                    </ul>
                  )}
                </div>
              </article>
            ))}
          </div>
        )}
      </section>
    </main>
  );
}
