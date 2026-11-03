export interface Ingredient {
  name: string;
  amount: number;
  unit: string;
}

export interface Recipe {
  id: string;
  created_at: string;
  updated_at: string;
  name: string;
  description: string;
  author: string;
  version: number;
  text: string;
  tags: string[];
  ingredients: Ingredient[];
  images: string[];
}

export async function getAllRecipes(): Promise<Recipe[]> {
  const response = await fetch('/api/v1/recipes');
  if (!response.ok) {
    throw new Error(`Rezepte konnten nicht geladen werden (HTTP ${response.status}).`);
  }
  return response.json() as Promise<Recipe[]>;
}
