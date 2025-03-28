import React from 'react';

function About() {
  return (
    <div className="about-container">
      <h1>About Vending Maxine</h1>
      
      <section className="about-section">
        <h2>Our Story</h2>
        <p>
          Vending Maxine was created to revolutionize the vending machine experience by 
          bringing modern technology and user-friendly interfaces to the traditional 
          vending experience.
        </p>
      </section>
      
      <section className="about-section">
        <h2>Our Mission</h2>
        <p>
          We aim to provide a seamless, convenient, and enjoyable vending experience 
          through intuitive digital interfaces and reliable service.
        </p>
      </section>
      
      <section className="about-section">
        <h2>Contact Us</h2>
        <p>
          Have questions or suggestions? Visit our GitHub repository:
          <br />
          <a href="https://github.com/zipizap/vendingMaxine" target="_blank" rel="noopener noreferrer">
            github.com/zipizap/vendingMaxine
          </a>
        </p>
      </section>
    </div>
  );
}

export default About;
