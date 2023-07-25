import { configureStore } from '@reduxjs/toolkit'
import postsReducer from './FeatureSlices/data'
import {TypedUseSelectorHook, useDispatch, useSelector} from "react-redux"


export const store = configureStore({
    reducer: {
        data: postsReducer,
       
    }, 
})


export type RootState = ReturnType<typeof store.getState>
export type AppDispatch = typeof store.dispatch

export const useAppDispatch: () => AppDispatch = useDispatch;
export const useAppSelector: TypedUseSelectorHook<RootState> = useSelector;