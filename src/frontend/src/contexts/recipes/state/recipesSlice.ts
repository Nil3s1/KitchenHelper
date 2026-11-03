import { createAsyncThunk, createSlice } from '@reduxjs/toolkit';
import { getAllRecipes, type Recipe } from '../api/recipesApi';

type LoadStatus = 'idle' | 'loading' | 'succeeded' | 'failed';

interface RecipesState {
  items: Recipe[];
  status: LoadStatus;
  error: string | null;
}

const initialState: RecipesState = {
  items: [],
  status: 'idle',
  error: null,
};

export const fetchRecipes = createAsyncThunk('recipes/fetchAll', async () => getAllRecipes());

const recipesSlice = createSlice({
  name: 'recipes',
  initialState,
  reducers: {},
  extraReducers: (builder) => {
    builder
      .addCase(fetchRecipes.pending, (state) => {
        state.status = 'loading';
        state.error = null;
      })
      .addCase(fetchRecipes.fulfilled, (state, action) => {
        state.status = 'succeeded';
        state.items = action.payload;
      })
      .addCase(fetchRecipes.rejected, (state, action) => {
        state.status = 'failed';
        state.error = action.error.message ?? 'Unbekannter Fehler beim Laden der Rezepte.';
      });
  },
});

export const recipesReducer = recipesSlice.reducer;
