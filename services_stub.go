package cinc

// Service type declarations. Methods are added in their own files/tasks.
type SearchService struct{ client *Client }
type CookbooksService struct{ client *Client }
type CookbookArtifactsService struct{ client *Client }
