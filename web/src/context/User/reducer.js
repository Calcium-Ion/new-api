export const reducer = (state, action) => {
  switch (action.type) {
    case 'login':
      return {
        ...state,
        user: action.payload,
      };
    case 'logout':
      return {
        ...state,
        user: undefined,
      };

    default:
      return state;
  }
};

export const initialState = {
  user: undefined,
};
