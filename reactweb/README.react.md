# Readme

*   **Vite Configuration:**  You can customize the build process and development server by modifying the `vite.config.ts` file in your project.
*   **TypeScript:**  This project uses TypeScript, which adds static typing to JavaScript.  This helps catch errors early and improves code maintainability.  Make sure you have a good understanding of TypeScript syntax and concepts.
*   **React:** This project uses React, a popular JavaScript library for building user interfaces. Familiarize yourself with React components, JSX syntax, and the React lifecycle.

## How to bootstrap a new React TypeScript project using Vite

    ```bash
    npm create vite@latest my-app -- --template react-ts
    cd my-app
    # reads the `package.json` file and installs all the listed packages and libraries required for the project to run
    npm install
    ```

## How to develop the project

    ```bash
    # Start the development 
    #  a) http://127.0.0.1:8080 golang webserver
    #  b) http://127.0.0.1:5173  vite webserver
    #     - hotreload: Vite will watch your files for changes and automatically reload the browser, making development faster.
    #     - proxy-forward certain paths to the golang webserver - see `vite.config.ts` for details
    # 
    npm run dev

    # edit src/* 
    # main entry point is usually `src/main.tsx` or `src/App.tsx`
    ```

## How to compile the project for deployment

    ```bash
    # Build the project for production, into the `dist` directory, 
    # The contents of the `dist` directory can be deployed to any static web server,
    npm run build

    ```

