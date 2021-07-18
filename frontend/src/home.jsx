const AUTH0_CLIENT_ID = "ceexXJ6MIeRzaqzbGNnlPK0tkLyygfx0";
const AUTH0_DOMAIN = "mtmiec.us.auth0.com";
const AUTH0_CALLBACK_URL = location.href;
const AUTH0_API_AUDIENCE = "webapp0.mattmiec.com";
import auth0 from 'auth0-js';
import React from 'react';

export class Home extends React.Component {
  constructor(props) {
    super(props);
    this.authenticate = this.authenticate.bind(this);
  }
  authenticate() {
    this.WebAuth = new auth0.WebAuth({
      domain: AUTH0_DOMAIN,
      clientID: AUTH0_CLIENT_ID,
      scope: "openid email",
      audience: AUTH0_API_AUDIENCE,
      responseType: "token id_token",
      redirectUri: AUTH0_CALLBACK_URL
    });
    this.WebAuth.authorize();
  }

  render() {
    return (
      <div className="container">
        <div className="row">
          <div className="col-sm-8 offset-sm-2 jumbotron text-center">
            <h1>Jokeish</h1>
            <p>A load of Dad jokes XD</p>
            <p>Sign in to get access </p>
            <button
              onClick={this.authenticate}
              className="btn btn-primary btn-lg btn-block"
            >
              Sign In
            </button>
          </div>
        </div>
      </div>
    );
  }
}
